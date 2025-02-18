package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Tabsync struct {
	isDead     []bool
	isTLS      bool
	isJitter   bool
	tabCounter int
}

type Globsync struct {
	outtext *widget.Entry
}

var glob Globsync

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())
	myWindow := myApp.NewWindow("Guatá")
	myWindow.Resize(fyne.NewSize(800, 800))
	tabContainer := container.NewAppTabs()
	tabContainer.SetTabLocation(container.TabLocationTop)
	var tabs Tabsync

	tabs.tabCounter = 0
	settingsWindow := myApp.NewWindow("Configurations")
	tlsContent := container.NewVBox(
		widget.NewCheck("TLS certificate/chain verify", func(isSet bool) { tabs.isTLS = isSet }),
	)
	pluginsContent := container.NewVBox(
		widget.NewLabel("Config de Plugins"),
	)
	mixContent := container.NewVBox(
		widget.NewCheck("Print req/resp jitter ", func(isSet bool) { tabs.isJitter = isSet }),
	)
	tabs_conf := container.NewAppTabs(
		container.NewTabItem("TLS", tlsContent),
		container.NewTabItem("Plugins", pluginsContent),
		container.NewTabItem("Others", mixContent),
	)
	settingsWindow.SetContent(tabs_conf)
	settingsWindow.SetFixedSize(true)
	settingsWindow.Resize(fyne.NewSize(350, 200))
	settingsWindow.CenterOnScreen()
	settingsWindow.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEscape {
			settingsWindow.Hide()
		}
	})
	settingsWindow.SetCloseIntercept(func() {
		settingsWindow.Hide()
	})

	settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		settingsWindow.Show()
	})
	addTabButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		createNewTab(&tabs, tabContainer, &myWindow)
	})
	removeTabButton := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		removeTab(tabContainer, &tabs.isDead)
	})
	buttonsContainer := container.NewHBox(addTabButton, removeTabButton, settingsButton)
	outputText := widget.NewMultiLineEntry()
	outputText.SetPlaceHolder("Guatá output!")
	outputText.Wrapping = fyne.TextWrapBreak
	outputText.Disable()
	glob.outtext = outputText
	accordion := widget.NewAccordion(
		widget.NewAccordionItem("", outputText),
	)
	myWindow.SetContent(container.NewBorder(buttonsContainer, accordion, nil, nil, tabContainer))
	createNewTab(&tabs, tabContainer, &myWindow)
	myWindow.CenterOnScreen()
	myWindow.SetOnClosed(func() {
		for i := range tabs.isDead {
			tabs.isDead[i] = true
		}
		settingsWindow.Close()
	})
	myWindow.ShowAndRun()
}
