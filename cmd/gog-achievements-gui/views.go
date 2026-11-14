package main

import (
	"fmt"
	"image"
	"math"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/timendum/gog-achievements/internal"
)

func panel(radius unit.Dp, w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		rec := op.Record(gtx.Ops)
		dims := w(gtx)
		call := rec.Stop()

		r := gtx.Dp(radius)
		cl := clip.RRect{
			Rect: image.Rectangle{Max: dims.Size},
			NE:   r, NW: r, SE: r, SW: r,
		}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, colPaper2)
		cl.Pop()
		call.Add(gtx.Ops)
		return dims
	}
}

func authStrip(th *material.Theme, s *appState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceBetween,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				title := material.H6(th, "GOG Achievements")
				title.Color = colInk
				title.TextSize = unit.Sp(18)
				return title.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Galaxy token not loaded"
				if s.auth != nil && s.auth.UserID != "" {
					label = fmt.Sprintf("Signed in · user %s", s.auth.UserID)
				}
				t := material.Caption(th, label)
				t.Color = colInk2
				return t.Layout(gtx)
			}),
		)
	}
}

func trustLine(th *material.Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		t := material.Caption(th, "Open source · reads GOG Galaxy refresh token locally · talks only to GOG / GOGDB · no telemetry")
		t.Color = colMuted
		return t.Layout(gtx)
	}
}

func searchField(th *material.Theme, s *appState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := material.Caption(th, "Search your library")
				label.Color = colInk2
				return label.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(panel(unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(th, &s.search, "Type a game title or ID…")
				ed.Color = colInk
				ed.HintColor = colMuted
				return layout.UniformInset(unit.Dp(10)).Layout(gtx, ed.Layout)
			})),
		)
	}
}

func accentButton(th *material.Theme, click *widget.Clickable, label string, enabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if !enabled {
			gtx = gtx.Disabled()
		}
		btn := material.Button(th, click, label)
		btn.Background = colAccent
		btn.Color = colAccentInk
		btn.CornerRadius = unit.Dp(8)
		btn.TextSize = unit.Sp(13)
		return btn.Layout(gtx)
	}
}

func ghostButton(th *material.Theme, click *widget.Clickable, label string, enabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if !enabled {
			gtx = gtx.Disabled()
		}
		btn := material.Button(th, click, label)
		btn.Background = colPaper3
		btn.Color = colInk
		btn.CornerRadius = unit.Dp(8)
		btn.TextSize = unit.Sp(13)
		return btn.Layout(gtx)
	}
}

func libraryView(th *material.Theme, w *app.Window, s *appState, invalidate func()) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		games := s.filteredGames()
		const (
			cardWidthDp = 210
			cardGapDp   = 12
			rowGapDp    = 12
		)
		cardWpx := gtx.Dp(unit.Dp(cardWidthDp))
		cardGapPx := gtx.Dp(unit.Dp(cardGapDp))
		avail := gtx.Constraints.Max.X
		if avail < 1 {
			avail = 1
		}
		cols := int(float32(avail+cardGapPx) / float32(cardWpx+cardGapPx))
		if cols < 1 {
			cols = 1
		}
		n := len(games)
		rows := int(math.Ceil(float64(n) / float64(cols)))
		if rows < 1 {
			rows = 1
		}

		s.libraryList.Axis = layout.Vertical
		list := material.List(th, &s.libraryList)
		list.AnchorStrategy = material.Overlay

		if n == 0 {
			msg := material.Body1(th, "No games match that search.")
			msg.Color = colInk2
			return msg.Layout(gtx)
		}

		return list.Layout(gtx, rows, func(gtx layout.Context, row int) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					children := make([]layout.FlexChild, 0, cols*2)
					for c := 0; c < cols; c++ {
						idx := row*cols + c
						cc := c
						if idx >= n {
							children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = cardWpx
								gtx.Constraints.Max.X = cardWpx
								return layout.Dimensions{Size: image.Pt(cardWpx, 0)}
							}))
						} else {
							it := games[idx]
							children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = cardWpx
								gtx.Constraints.Max.X = cardWpx
								return gameCard(th, w, it, func(id int, title string) {
									s.openGame(id, title, invalidate)
								})(gtx)
							}))
						}
						if cc != cols-1 {
							children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(cardGapDp)}.Layout(gtx)
							}))
						}
					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: unit.Dp(rowGapDp)}.Layout(gtx)
				}),
			)
		})
	}
}

func handleAchievementChrome(gtx layout.Context, s *appState, invalidate func()) {
	if s.backBtn.Clicked(gtx) && !s.busy {
		s.view = viewLibrary
		s.status = ""
		s.statusIsError = false
	}
	if s.selectLockedBtn.Clicked(gtx) && !s.busy {
		s.setSelection(func(a internal.Achievement) bool { return a.DateUnlocked == nil })
	}
	if s.selectUnlockBtn.Clicked(gtx) && !s.busy {
		s.setSelection(func(a internal.Achievement) bool { return a.DateUnlocked != nil })
	}
	if s.clearSelectBtn.Clicked(gtx) && !s.busy {
		s.clearSelection()
	}
	if s.unlockBtn.Clicked(gtx) && !s.busy {
		s.runBulk(true, invalidate)
	}
	if s.lockBtn.Clicked(gtx) && !s.busy {
		s.runBulk(false, invalidate)
	}
}

func achievementsView(th *material.Theme, s *appState, invalidate func()) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		handleAchievementChrome(gtx, s, invalidate)
		selCount, _, _ := s.selectedCount()
		busy := s.busy

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(ghostButton(th, &s.backBtn, "← Library", !busy)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						t := material.H6(th, s.selectedTitle)
						t.Color = colInk
						t.TextSize = unit.Sp(16)
						t.MaxLines = 1
						return t.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(accentButton(th, &s.unlockBtn, "Unlock selected", !busy && selCount > 0)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(ghostButton(th, &s.lockBtn, "Lock selected", !busy && selCount > 0)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(ghostButton(th, &s.selectLockedBtn, "Select locked", !busy)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(ghostButton(th, &s.selectUnlockBtn, "Select unlocked", !busy)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(ghostButton(th, &s.clearSelectBtn, "Clear", !busy && selCount > 0)),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if s.status == "" {
					hint := material.Caption(th, fmt.Sprintf("%d selected", selCount))
					hint.Color = colMuted
					return hint.Layout(gtx)
				}
				t := material.Caption(th, s.status)
				if s.statusIsError {
					t.Color = colDanger
				} else {
					t.Color = colSuccess
				}
				return t.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if s.achLoading {
					t := material.Body1(th, "Loading achievements…")
					t.Color = colInk2
					return t.Layout(gtx)
				}
				if s.achErr != "" {
					t := material.Body1(th, s.achErr)
					t.Color = colDanger
					return t.Layout(gtx)
				}
				if len(s.achievements) == 0 {
					t := material.Body1(th, "No achievements found for this game.")
					t.Color = colInk2
					return t.Layout(gtx)
				}
				s.achList.Axis = layout.Vertical
				list := material.List(th, &s.achList)
				return list.Layout(gtx, len(s.achievements), func(gtx layout.Context, i int) layout.Dimensions {
					return achievementRow(th, s.achievements[i])(gtx)
				})
			}),
		)
	}
}

func achievementRow(th *material.Theme, it *achItem) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		unlocked := it.ach.DateUnlocked != nil
		return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, panel(unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						cb := material.CheckBox(th, &it.selected, "")
						cb.Color = colAccent
						return cb.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						size := gtx.Dp(unit.Dp(10))
						defer clip.Ellipse{Max: image.Pt(size, size)}.Push(gtx.Ops).Pop()
						if unlocked {
							paint.Fill(gtx.Ops, colSuccess)
						} else {
							paint.Fill(gtx.Ops, colMuted)
						}
						return layout.Dimensions{Size: image.Pt(size, size)}
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						name := it.ach.Name
						if strings.TrimSpace(name) == "" {
							name = it.ach.AchievementKey
						}
						status := "Locked"
						if unlocked {
							status = "Unlocked"
						}
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								t := material.Body1(th, name)
								t.Color = colInk
								t.TextSize = unit.Sp(14)
								t.MaxLines = 1
								return t.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								desc := it.ach.Description
								if desc == "" {
									desc = status
								} else {
									desc = status + " · " + desc
								}
								t := material.Caption(th, desc)
								t.Color = colInk2
								t.MaxLines = 2
								return t.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								t := material.Caption(th, "ID "+it.ach.AchievementID)
								t.Color = colMuted
								return t.Layout(gtx)
							}),
						)
					}),
				)
			})
		}))
	}
}
