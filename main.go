package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())
	myWindow := myApp.NewWindow("Requisição TCP Cliente")
	myWindow.Resize(fyne.NewSize(800, 600))	
	tabContainer := container.NewAppTabs()
	tabContainer.SetTabLocation(container.TabLocationTop)

	var tabCounter int
	var mu sync.Mutex

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("host.com:8080")
	hostEntry.Wrapping = fyne.TextTruncate
	hostEntryContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, hostEntry.MinSize().Height)), hostEntry)

	requestText := widget.NewMultiLineEntry()
	requestText.SetPlaceHolder("Raw HTTP/1.1 request")
	requestText.Wrapping = fyne.TextWrapBreak

	responseText := widget.NewMultiLineEntry()
	responseText.SetPlaceHolder("Raw HTTP/1.1 response")
	responseText.Disable()
	responseText.Wrapping = fyne.TextWrapBreak

	sendButton := widget.NewButton("Send", func() {
		host := hostEntry.Text
		request := requestText.Text		
		response, err := sendTCPRequest(host, request)
		if err != nil {
			responseText.SetText(fmt.Sprintf("Error: %v", err))
			return
		}
		responseText.SetText(string(response))		
	})

	sendButtonContainer := container.New(layout.NewGridWrapLayout(sendButton.MinSize()), sendButton)

	textSplitContainer := container.NewHSplit(requestText, responseText)
	textSplitContainer.SetOffset(0.5)
	textSplitWrapper := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(800, 500)),
		textSplitContainer,
	)

	createNewTab := func() {
		mu.Lock()
		defer mu.Unlock()
		tabCounter++

		content := container.NewVBox(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(0, 10))),
			container.NewHBox(
				widget.NewLabel("Host:Port "),
				hostEntryContainer,
				sendButtonContainer,
			),
			widget.NewLabel("Request:"),
			textSplitWrapper,
		)

		tabContainer.Append(container.NewTabItem(fmt.Sprintf("Tab %d", tabCounter), content))		
	}

	removeTab := func() {
		mu.Lock()
		defer mu.Unlock()

		if len(tabContainer.Items) > 0 {
			selectedTab := tabContainer.Selected()

			tabIndex := -1
			for i, item := range tabContainer.Items {
				if item == selectedTab {
					tabIndex = i
					break
				}
			}

			if tabIndex != -1 {
				copy(tabContainer.Items[tabIndex:], tabContainer.Items[tabIndex+1:])
				tabContainer.Items = tabContainer.Items[:len(tabContainer.Items)-1]

				if len(tabContainer.Items) > 0 {
					if tabIndex < len(tabContainer.Items) {
						tabContainer.SelectTabIndex(tabIndex)
					} else {
						tabContainer.SelectTabIndex(len(tabContainer.Items) - 1)
					}
				}
			}
		}
		tabContainer.Refresh()
	}

	addTabButton := container.New(layout.NewCenterLayout(), widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		createNewTab()
	}))
	removeTabButton := container.New(layout.NewCenterLayout(), widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		removeTab()
	}))	

	ButtonsContainer := container.NewHBox(addTabButton, removeTabButton)

	myWindow.SetContent(
		container.NewVBox(ButtonsContainer, tabContainer),
	)
	createNewTab()

	myWindow.ShowAndRun()
}

func sendTCPRequest(host string, request string) ([]byte, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(request))
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}

	resp, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler a resposta: %w", err)
	}

	return resp, nil
}
