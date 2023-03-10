package pdf

import (
	"bytes"
	"context"
	"io"
	"pdfserver/chrome"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Chrome struct {
	chrome *chrome.Chrome
}

func NewChrome(chrome *chrome.Chrome) *Chrome {
	return &Chrome{chrome: chrome}
}

func (c *Chrome) Render(ctx context.Context, w io.Writer, content string) error {
	b, err := c.SaveAsPDF(ctx, content)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

func (c *Chrome) SaveAsPDF(ctx context.Context, content string) ([]byte, error) {
	var buf []byte

	err := c.chrome.Run(ctx, chromedp.Tasks{
		enableLifeCycleEvents(),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return page.SetDocumentContent(cdp.FrameID(chromedp.FromContext(ctx).Target.TargetID.String()), content).Do(ctx)
		}),
		waitUntil("networkIdle"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				Do(ctx)
			return err
		}),
	})
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}
		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

func waitUntil(eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		ch := make(chan struct{})
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch e := ev.(type) {
			case *page.EventLifecycleEvent:
				if e.Name == eventName {
					close(ch)
				}
			}
		})
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
