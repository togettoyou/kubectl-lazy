package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ui struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	client     *client
	app        *tview.Application
	namespace  string
	namespaces *tview.List
	pod        string
	pods       *tview.List
	tabPage    *tabPage
}

var content = []string{"\nInfos", "Events", "Logs"}

type tabPage struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	tab        string
	tabs       *tview.TextView
	contents   *tview.TextView
}

func (u *ui) clearNamespaces() {
	if u == nil {
		return
	}
	u.namespace = ""
	u.namespaces.Clear()
}

func (u *ui) clearPods() {
	if u == nil {
		return
	}
	if u.cancelFunc != nil {
		u.cancelFunc()
	}
	u.ctx, u.cancelFunc = context.WithCancel(context.Background())
	u.pod = ""
	u.pods.Clear()
}

func (tp *tabPage) clear() {
	if tp.cancelFunc != nil {
		tp.cancelFunc()
	}
	tp.ctx, tp.cancelFunc = context.WithCancel(context.Background())
	tp.contents.Clear()
}

func NewUi(client *client) *ui {
	u := &ui{
		client: client,
		app:    tview.NewApplication(),
	}
	u.initView()
	return u
}

func (u *ui) Run() error {
	return u.app.Run()
}

func (u *ui) initView() {
	u.initTabPages()
	u.initPods()
	u.initNamespaces()
	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(u.namespaces, 0, 1, true).
			AddItem(u.pods, 0, 2, false),
			0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(u.tabPage.tabs, 0, 1, false).
			AddItem(u.tabPage.contents, 0, 5, false),
			0, 2, true)
	u.app.SetRoot(flex, true).EnableMouse(true).SetFocus(flex)
}

func (u *ui) initNamespaces() {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle("namespaces")
	list.SetFocusFunc(func() {
		if u.namespace != "" {
			return
		}
		u.clearNamespaces()
		u.clearPods()
		u.tabPage.clear()

		namespaces, err := u.client.Namespaces(u.ctx)
		if err != nil {
			return
		}
		for _, namespace := range namespaces {
			list.AddItem(namespace, "", 0, nil)
		}
	})
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		u.clearPods()
		u.tabPage.clear()

		u.namespace = mainText

		go func() {
			pods, err := u.client.Pods(u.ctx, u.namespace)
			if err != nil {
				return
			}
			for _, pod := range pods {
				u.pods.AddItem(pod, "", 0, nil)
				u.app.Draw()
			}
		}()

	})
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			u.app.SetFocus(u.pods)
		}
		return event
	})
	u.namespaces = list
}

func (u *ui) initPods() {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle("pods")
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		u.tabPage.clear()

		u.pod = mainText

		u.updateTabPageContents()
	})
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBackspace2:
			u.app.SetFocus(u.namespaces)
		case tcell.KeyEnter:
			u.app.SetFocus(u.tabPage.tabs)
		}
		return event
	})
	u.pods = list
}

func (u *ui) initTabPages() {
	contents := tview.NewTextView().SetText(content[0])
	contents.SetBorder(true).SetTitle(content[0])
	contents.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBackspace2:
			u.app.SetFocus(u.tabPage.tabs)
		}
		return event
	})

	tabs := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			if u.tabPage == nil {
				return
			}
			u.tabPage.clear()
			index, err := strconv.Atoi(added[0])
			if err != nil {
				return
			}
			u.tabPage.tab = content[index]
			u.tabPage.contents.SetTitle(content[index])
			u.updateTabPageContents()
		})
	tabs.SetTitle("[Enter 确认] [Backspace 回退] [↑ ↓ 切换] [Ctrl C 退出]").SetBorder(true)

	previousSlide := func() {
		slide, _ := strconv.Atoi(tabs.GetHighlights()[0])
		slide = (slide - 1 + len(content)) % len(content)
		tabs.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	nextSlide := func() {
		slide, _ := strconv.Atoi(tabs.GetHighlights()[0])
		slide = (slide + 1) % len(content)
		tabs.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	for index, name := range content {
		fmt.Fprintf(tabs, `["%d"][green]%s[white][""]  `, index, name)
	}
	tabs.Highlight("0")
	tabs.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBackspace2:
			u.app.SetFocus(u.pods)
		case tcell.KeyRight:
			nextSlide()
			return nil
		case tcell.KeyLeft:
			previousSlide()
			return nil
		case tcell.KeyEnter:
			u.app.SetFocus(u.tabPage.contents)
		}
		return event
	})

	u.tabPage = &tabPage{
		tab:      content[0],
		tabs:     tabs,
		contents: contents,
	}
}

func (u *ui) updateTabPageContents() {
	go func() {
		switch u.tabPage.tab {
		case content[0]:
			infos, err := u.client.Infos(u.tabPage.ctx, u.namespace, u.pod)
			if err != nil {
				return
			}
			infosData, err := json.MarshalIndent(&infos, "", "  ")
			if err != nil {
				return
			}
			u.tabPage.contents.SetText(string(infosData))
			u.app.Draw()
		case content[1]:
			events, err := u.client.Events(u.tabPage.ctx, u.namespace, u.pod)
			if err != nil {
				return
			}
			eventsData, err := json.MarshalIndent(&events, "", "  ")
			if err != nil {
				return
			}
			u.tabPage.contents.SetText(string(eventsData))
			u.app.Draw()
		case content[2]:
			logs, err := u.client.Logs(u.tabPage.ctx, u.namespace, u.pod)
			if err != nil {
				return
			}
			for log := range logs {
				fmt.Fprintf(u.tabPage.contents, log)
				u.app.Draw()
			}
		}
	}()
}
