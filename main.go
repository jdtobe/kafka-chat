package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/segmentio/kafka-go"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "test-chat"

	username string

	reader *kafka.Reader
	writer *kafka.Writer
)

type message struct {
	t    time.Time
	From string
	To   string
	Body string
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("You must provide your username as the first argument when running ", os.Args[0])
	}
	username = os.Args[1]

	if err := ui.Init(); err != nil {
		log.Fatalf("Unable to initialize termui: %v", err)
	}
	defer ui.Close()

	mWidget := widgets.NewParagraph()
	mWidget.Title = "Messages"
	mWidget.TitleStyle.Fg = ui.ColorWhite
	mWidget.BorderStyle.Fg = ui.ColorBlue

	iWidget := widgets.NewParagraph()
	iWidget.Title = "Input:"
	iWidget.TextStyle.Fg = ui.ColorGreen
	iWidget.BorderStyle.Fg = ui.ColorYellow

	resize := func(w, h int) {
		mWidget.SetRect(0, 0, w, h-3)
		iWidget.SetRect(0, h-3, w, h)
	}
	resize(ui.TerminalDimensions())
	ui.Render(mWidget, iWidget)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go subscribe(ctx, mWidget)
	go produce(ctx, &wg, iWidget, resize)

	wg.Wait()
	cancel()
}

func subscribe(ctx context.Context, p *widgets.Paragraph) {
	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	defer reader.Close()
	reader.SetOffset(-1) // start at the end

	// offset := int64(0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Fatalf("Unable to fetch message: %v", err)
		}
		// offset = msg.Offset

		var m message
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Fatalf("Unable to unmarshal message: %v", err)
		}

		if m.From != username {
			p.Text += fmt.Sprintf("%s: %s\n", m.From, m.Body)
			ui.Render(p)
		}
	}
}

func produce(ctx context.Context, wg *sync.WaitGroup, p *widgets.Paragraph, resizeFn func(w, h int)) {
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	var line string
	evtCh := ui.PollEvents()
	for event := range evtCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if event.Type == ui.ResizeEvent {
			resizeFn(ui.TerminalDimensions())
			continue
		}
		if event.Type != ui.KeyboardEvent {
			continue
		}

		switch event.ID {
		case "<C-c>":
			os.Exit(0)
		case "<Resize>":
			resizeFn(ui.TerminalDimensions())
		case "<Backspace>":
			l := len(line)
			if l == 0 {
				continue
			}
			line = line[:l-1]
			p.Text = line
			ui.Render(p)
		case "<Enter>":
			if line == "/quit" {
				break
			}

			m := message{
				t:    time.Now(),
				From: username,
				To:   "anyone",
				Body: line,
			}
			msg, err := json.Marshal(m)
			if err != nil {
				log.Fatalf("Unable to marshal output message: %v", err)
			}

			err = writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte("message"),
				Value: msg,
			})
			if err != nil {
				log.Fatalf("Unable to write message: %v", err)
			}

			p.Text = ""
			ui.Render(p)
		default:
			if event.ID == "<Space>" {
				event.ID = " "
			}
			line = line + event.ID
			p.Text = line
			ui.Render(p)
			if event.ID != "\n" {
				continue // not the end of the line, get more chars
			}
		}
	}

	wg.Done()
}
