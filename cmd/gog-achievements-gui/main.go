package main

import (
	"fmt"
	"image"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/timendum/gog-achievements/internal"
)

func main() {
	go runUI()
	app.Main()
}

func runUI() {
	w := new(app.Window)
	w.Option(app.Title("GOG Achievements Manager"))
	w.Option(app.Size(unit.Dp(1100), unit.Dp(740)))

	refreshToken, err := internal.GetRefreshToken()
	if err != nil {
		fatalWindow(w, fmt.Sprintf("Could not read the GOG Galaxy refresh token from the Windows registry.\n\n%v\n\nOpen GOG Galaxy, sign in, then try again.", err))
		return
	}
	if refreshToken == "" {
		fatalWindow(w, "No GOG Galaxy refresh token was found.\n\nOpen GOG Galaxy, sign in, then try again.")
		return
	}

	authResp, err := internal.GetAuth(refreshToken, "", "")
	if err != nil {
		fatalWindow(w, fmt.Sprintf("Authentication with GOG failed.\n\n%v", err))
		return
	}
	if authResp == nil || authResp.RefreshToken == "" {
		fatalWindow(w, "Authentication response was empty.\n\nSign in to GOG Galaxy and try again.")
		return
	}

	owned := internal.ListOwnedGameIDs(*authResp)
	if owned == nil {
		fatalWindow(w, "Could not load your owned games from GOG.\n\nCheck your network connection and try again.")
		return
	}

	state := newAppState(authResp, refreshToken, *owned)
	th := newTheme()
	var ops op.Ops
	invalidate := func() { w.Invalidate() }

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			if e.Err != nil {
				log.Println("window error:", e.Err)
			}
			os.Exit(0)
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			paint.Fill(gtx.Ops, colPaper)

			layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(authStrip(th, state)),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						h := gtx.Dp(unit.Dp(1))
						defer clip.Rect{Max: image.Pt(gtx.Constraints.Max.X, h)}.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, colRule)
						return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, h)}
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						switch state.view {
						case viewAchievements:
							return achievementsView(th, state, invalidate)(gtx)
						default:
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(searchField(th, state)),
								layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
								layout.Flexed(1, libraryView(th, w, state, invalidate)),
							)
						}
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
					layout.Rigid(trustLine(th)),
				)
			})
			e.Frame(gtx.Ops)
		}
	}
}

func fatalWindow(w *app.Window, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	th := newTheme()
	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			os.Exit(1)
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			paint.Fill(gtx.Ops, colPaper)
			layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						t := material.H6(th, "GOG Achievements")
						t.Color = colInk
						return t.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						t := material.Body1(th, msg)
						t.Color = colInk2
						return t.Layout(gtx)
					}),
				)
			})
			e.Frame(gtx.Ops)
		}
	}
}
