package main

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/timendum/gog-achievements/internal"

	_ "golang.org/x/image/webp"
)

type gameItem struct {
	id    int
	click widget.Clickable

	mu            sync.RWMutex
	title         string
	detailLoaded  bool
	detailLoading bool
	imageURL      string
	imgLoaded     bool
	imgLoading    bool
	imgOp         paint.ImageOp
}

func newGameItem(id int) *gameItem {
	return &gameItem{id: id, title: fmt.Sprintf("%d", id)}
}

func (g *gameItem) Title() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.title
}

func (g *gameItem) HasImage() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.imgLoaded
}

func (g *gameItem) ImageOp() paint.ImageOp {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.imgOp
}

var (
	detailsSem = make(chan struct{}, 6)
	imagesSem  = make(chan struct{}, 8)
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

func (g *gameItem) loadGameDetail(w *app.Window, get func(int) *internal.GameDetail) {
	g.mu.Lock()
	if g.detailLoaded || g.detailLoading {
		g.mu.Unlock()
		return
	}
	g.detailLoading = true
	g.mu.Unlock()

	go func(id int) {
		detailsSem <- struct{}{}
		defer func() { <-detailsSem }()

		d := get(id)
		if d == nil {
			g.mu.Lock()
			g.detailLoading = false
			g.detailLoaded = true
			g.mu.Unlock()
			w.Invalidate()
			return
		}

		g.mu.Lock()
		g.title = d.Title
		if d.ImageLogo != "" {
			g.imageURL = fmt.Sprintf("https://images.gog-statics.com/%s_product_tile_extended_432x243.webp", d.ImageLogo)
		}
		g.detailLoaded = true
		g.detailLoading = false
		g.mu.Unlock()
		w.Invalidate()
		if g.imageURL != "" {
			g.downloadPrintImage(w)
		}
	}(g.id)
}

func (g *gameItem) downloadPrintImage(w *app.Window) {
	g.mu.Lock()
	if g.imgLoaded || g.imgLoading || g.imageURL == "" {
		g.mu.Unlock()
		return
	}
	url := g.imageURL
	g.imgLoading = true
	g.mu.Unlock()

	go func(u string) {
		imagesSem <- struct{}{}
		defer func() { <-imagesSem }()

		img, err := fetchImage(u)
		g.mu.Lock()
		defer g.mu.Unlock()
		g.imgLoading = false
		if err != nil {
			w.Invalidate()
			return
		}
		g.imgOp = paint.NewImageOp(img)
		g.imgLoaded = true
		w.Invalidate()
	}(url)
}

func fetchImage(url string) (image.Image, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("image status %d", resp.StatusCode)
	}
	img, _, err := image.Decode(resp.Body)
	return img, err
}

func gameCard(th *material.Theme, w *app.Window, it *gameItem, onOpen func(id int, title string)) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		it.loadGameDetail(w, internal.GetGameDetail)
		it.downloadPrintImage(w)

		for it.click.Clicked(gtx) {
			onOpen(it.id, it.Title())
		}

		return it.click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			r := gtx.Dp(unit.Dp(10))
			defer clip.RRect{
				Rect: image.Rectangle{Max: gtx.Constraints.Max},
				NE:   r, NW: r, SE: r, SW: r,
			}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, colPaper2)

			return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						h := gtx.Dp(unit.Dp(110))
						gtx.Constraints.Min.Y = h
						gtx.Constraints.Max.Y = h
						rr := gtx.Dp(unit.Dp(6))
						defer clip.RRect{
							Rect: image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, h)},
							NE:   rr, NW: rr, SE: rr, SW: rr,
						}.Push(gtx.Ops).Pop()
						if it.HasImage() {
							img := widget.Image{Src: it.ImageOp(), Fit: widget.Cover}
							return img.Layout(gtx)
						}
						paint.Fill(gtx.Ops, colPaper3)
						return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, h)}
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						t := material.Body1(th, it.Title())
						t.Color = colInk
						t.MaxLines = 2
						t.TextSize = unit.Sp(14)
						return t.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						t := material.Caption(th, fmt.Sprintf("ID %d", it.id))
						t.Color = colMuted
						return t.Layout(gtx)
					}),
				)
			})
		})
	}
}
