package chrome

import (
	"context"

	"github.com/chromedp/chromedp"
	"golang.org/x/sync/semaphore"
)

type Chrome struct {
	ctx context.Context
	sem *semaphore.Weighted
}

func NewChrome(maxTab int64, opts ...chromedp.ExecAllocatorOption) (*Chrome, error) {
	if maxTab == 0 {
		maxTab = 20
	}

	browserCtx, _ := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], opts...)...)
	ctx, _ := chromedp.NewContext(browserCtx)
	if err := chromedp.Run(ctx); err != nil {
		return nil, err
	}

	return &Chrome{
		sem: semaphore.NewWeighted(maxTab),
		ctx: ctx,
	}, nil
}

func (c *Chrome) Close() error {
	if c.ctx == nil {
		return nil
	}
	return chromedp.Cancel(c.ctx)
}

func (c *Chrome) Run(ctx context.Context, actions ...chromedp.Action) error {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer c.sem.Release(1)

	tabCtx, cancel := chromedp.NewContext(c.ctx)
	defer cancel()

	return chromedp.Run(tabCtx, actions...)
}
