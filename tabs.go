package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type insync struct {
	synDead   *bool
	syncAbort bool
	tabs      *Tabsync
	id        int
}

func createNewTab(tabs *Tabsync, tabContainer *container.AppTabs, w *fyne.Window) {
	tabs.tabCounter++
	tabs.isDead = append(tabs.isDead, false)
	var lsync insync
	lsync.synDead = &tabs.isDead[tabs.tabCounter-1]
	lsync.syncAbort = false
	lsync.tabs = tabs
	lsync.id = tabs.tabCounter

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("localhost:8080")
	hostEntry.Wrapping = fyne.TextTruncate
	hostEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, hostEntry.MinSize().Height)), hostEntry)

	requestText := widget.NewMultiLineEntry()
	requestText.SetPlaceHolder("Raw HTTP/1.1 request")
	requestText.Wrapping = fyne.TextWrapBreak
	requestText.Resize(fyne.NewSize(400, 450))
	//requestContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(400,600)), requestText)

	responseText := widget.NewMultiLineEntry()
	responseText.SetPlaceHolder("Raw HTTP/1.1 response")
	responseText.Wrapping = fyne.TextWrapBreak
	responseText.Resize(fyne.NewSize(400, 450))
	//responseContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(400,600)), responseText)

	// Botão de envio
	sendButton := widget.NewButton("Send", nil)
	sendButton.OnTapped = sendFunc(&lsync, hostEntry, requestText, responseText, sendButton)

	// Layout  Host e Send
	topRow :=
		container.NewHBox(
			widget.NewLabel("Host:Port"),
			hostEntryContainer,
			container.New(layout.NewGridWrapLayout(fyne.NewSize(2, 0))),
			sendButton)

	textSplitContainer := container.NewGridWithColumns(2, requestText, responseText)

	content := container.NewVBox(
		widget.NewSeparator(),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(0, 5))),
		topRow,
		container.New(layout.NewMaxLayout(), textSplitContainer),
	)
	tabContainer.Append(container.NewTabItem("Tab "+strconv.Itoa(tabs.tabCounter), content))
	tabContainer.SelectIndex(tabs.tabCounter - 1)
	go monitorResize(&tabs.isDead[tabs.tabCounter-1], w, requestText, responseText)
}

// Removes tab, doesn't dealloc all resources
func removeTab(tabContainer *container.AppTabs, isDead *[]bool) {
	if len(tabContainer.Items) > 0 {
		selectedIndex := tabContainer.SelectedIndex()
		selectedTabItem := tabContainer.Items[selectedIndex]
		content := selectedTabItem.Content
		if c, ok := content.(*fyne.Container); ok {
			c.RemoveAll()
		}
		(*isDead)[selectedIndex] = true
		tabContainer.RemoveIndex(tabContainer.SelectedIndex())
	}
}
