package main

import (
	"crypto/tls"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	outtext   *widget.Entry
	tlsVerify *widget.Check
	othjitter *widget.Check
	mtls      *tls.Certificate
}

var glob Globsync

func mtlsfile(w fyne.Window, label *widget.Label, cert *tls.Certificate) {
	var passwordDialog *dialog.CustomDialog
	var er error
	fileDialog := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		filePath := r.URI().Path()
		myPrint("Selected cert: " + filePath)
		label.SetText(filePath)
		if filePath[len(filePath)-2:] == "fx" || filePath[len(filePath)-2:] == "12" {
			passwordEntry := widget.NewEntry()
			passwordEntry.SetPlaceHolder("No pass")
			okButton := widget.NewButton("OK", func() {
				password := passwordEntry.Text
				passwordDialog.Hide()
				myPrint("PKCS#12 mTLS mode - pass: " + password)
				*cert, er = loadPKCS12Certificate(filePath, password)
			})
			passwordDialog = dialog.NewCustom("Certificate Pass", "X", container.NewVBox(
				widget.NewLabel("PKCS#12 password:"),
				passwordEntry,
				okButton,
			), w)
			passwordDialog.Show()
		} else {
			myPrint("PEM mTLS mode")
			*cert, er = tls.LoadX509KeyPair(filePath, filePath)
		}
		if er != nil {
			myPrint(fmt.Sprintf("certificate error: %v\n", err))
		}
	}, w)

	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pem", ".crt", ".p12", ".pfx"}))
	fileDialog.Show()
}

func main() {
	myApp := app.NewWithID("br.guata")
	myApp.Settings().SetTheme(theme.DarkTheme())
	myWindow := myApp.NewWindow("Guatá")
	myWindow.Resize(fyne.NewSize(800, 800))
	tabContainer := container.NewAppTabs()
	tabContainer.SetTabLocation(container.TabLocationTop)
	var tabs Tabsync

	tabs.tabCounter = 0
	// Default configs
	settingsWindow := myApp.NewWindow("Configurations")
	glob.tlsVerify = widget.NewCheck("TLS certificate/chain verify", func(isSet bool) { tabs.isTLS = isSet })
	glob.othjitter = widget.NewCheck("Print req/resp jitter ", func(isSet bool) { tabs.isJitter = isSet })
	glob.othjitter.SetChecked(true)
	glob.mtls = &tls.Certificate{}
	mTLStxt := widget.NewLabel("")
	mTLSBtt := widget.NewButtonWithIcon("", theme.DocumentIcon(), func() {
		mtlsfile(myWindow, mTLStxt, glob.mtls)
	})
	tlsContent := container.NewVBox(
		container.NewHBox(
			mTLSBtt,
			widget.NewLabel("Client-cert (mTLS):"),
			mTLStxt,
		),
		glob.tlsVerify,
	)
	pluginsContent := container.NewVBox(
		widget.NewLabel("Config Plugins??"),
	)
	mixContent := container.NewVBox(
		glob.othjitter,
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
