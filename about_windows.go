//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	appVersion = "1.1.0"
	appSiteURL = "https://www.dwrean.net/"
	appRepoURL = "https://github.com/dwrean-net/dwrean-cleaner"
	idAbout = 1002
	idSite = 2001
	idRepo = 2002
	idDismiss = 2003
	wmSetIcon = 0x0080
	idiApplication = 32512
)

var (
	shell32 = syscall.NewLazyDLL("shell32.dll")
	pShellExecute = shell32.NewProc("ShellExecuteW")
	pMessageBox = user32.NewProc("MessageBoxW")
	pLoadImage = user32.NewProc("LoadImageW")
	pDestroyWindow = user32.NewProc("DestroyWindow")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	aboutWindow uintptr
	aboutClassRegistered bool
	aboutClassName = utf16("DwreanCleanerModernAboutV11")
	aboutTitleFont uintptr
)

func openOfficialURL(url string) {
	// Limit ShellExecute to the publisher's official, fixed destinations.
	if url!=appSiteURL && url!=appRepoURL {return}
	r:=call(pShellExecute,mainWindow,ptr("open"),ptr(url),0,0,1)
	if r<=32 {
		call(pMessageBox,mainWindow,
			ptr("Δεν ήταν δυνατό να ανοίξει ο σύνδεσμος."),
			ptr("dwrean Cleaner"),0x30)
	}
}

func paintAbout(hwnd uintptr) {
	var ps paintStruct
	hdc:=call(pBeginPaint,hwnd,uintptr(unsafe.Pointer(&ps)))
	if hdc==0 {return}
	fillRect(hdc,rect(0,0,555,426),cCanvas)
	fillRect(hdc,rect(0,0,555,5),cBlue)

	if appIcon!=0 {
		call(pDrawIconEx,hdc,31,27,appIcon,52,52,0,0,3)
	}
	drawText(hdc,"dwrean Cleaner",rect(94,26,413,36),fontLogo,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"PORTABLE UTILITY",rect(96,65,360,23),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)

	roundRect(hdc,rect(25,113,505,193),rgb(229,236,246),27)
	roundRect(hdc,rect(23,110,508,193),cWhite,27)

	drawText(hdc,"Πληροφορίες εφαρμογής",rect(44,128,452,31),aboutTitleFont,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"Έκδοση",rect(45,170,110,24),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,appVersion+"  (σταθερή)",rect(159,170,318,24),fontBody,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"Δημιουργός",rect(45,202,110,24),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"Κυριάκος Οικονομίδης",rect(159,202,319,24),fontBody,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"Ιστότοπος",rect(45,235,110,24),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"dwrean.net",rect(159,235,318,24),fontBody,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"Άδεια",rect(45,266,110,24),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"MIT  •  Δωρεάν  •  Open source",rect(159,266,325,24),fontBody,cNavy,dtLeft|dtVCenter|dtSingleLine)

	drawText(hdc,"Μάθε περισσότερα και βρες τον πηγαίο κώδικα:",rect(30,323,493,27),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"© 2026 dwrean.net",rect(29,393,300,23),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	call(pEndPaint,hwnd,uintptr(unsafe.Pointer(&ps)))
}

func aboutWndProc(hwnd uintptr,msg uint32,wparam,lparam uintptr) uintptr {
	switch msg {
	case wmPaint:
		paintAbout(hwnd)
		return 0
	case wmDrawItem:
		if lparam!=0 {
			item:=(*drawItem)(unsafe.Pointer(lparam))
			if item.CtlID==idSite || item.CtlID==idRepo || item.CtlID==idDismiss {
				drawModernButton(item)
				return 1
			}
		}
	case wmCommand:
		switch uint32(wparam&0xffff) {
		case idSite:
			openOfficialURL(appSiteURL)
			return 0
		case idRepo:
			openOfficialURL(appRepoURL)
			return 0
		case idDismiss:
			call(pDestroyWindow,hwnd)
			return 0
		}
	case wmClose:
		call(pDestroyWindow,hwnd)
		return 0
	case wmDestroy:
		aboutWindow=0
		return 0
	}
	return call(pDefWindow,hwnd,uintptr(msg),wparam,lparam)
}

func showAbout(){
	if aboutWindow!=0 {
		call(pSetForegroundWindow,aboutWindow)
		return
	}
	instance:=call(pGetModuleHandle,0)
	if !aboutClassRegistered {
		class:=wndClass{
			CbSize:uint32(unsafe.Sizeof(wndClass{})),
			WndProc:syscall.NewCallback(aboutWndProc),
			Instance:instance,Icon:appIcon,IconSmall:appIcon,
			Cursor:call(pLoadCursor,0,32512),
			Background:backgroundBrush,ClassName:aboutClassName,
		}
		if call(pRegisterClass,uintptr(unsafe.Pointer(&class)))==0 {return}
		aboutClassRegistered=true
	}
	if aboutTitleFont==0 {aboutTitleFont=font(20,700)}
	aboutWindow=call(pCreateWindow,0,uintptr(unsafe.Pointer(aboutClassName)),
		ptr("Περί • dwrean Cleaner"),
		wsCaption|wsSysMenu,
		0x80000000,0x80000000,555,472,mainWindow,0,instance,0)
	if aboutWindow==0 {return}
	button(aboutWindow,idSite,"dwrean.net",rect(26,353,237,43))
	button(aboutWindow,idRepo,"GitHub",rect(274,353,158,43))
	button(aboutWindow,idDismiss,"Κλείσιμο",rect(441,353,87,43))
	call(pShowWindow,aboutWindow,swShow)
	call(pUpdateWindow,aboutWindow)
}

func loadAppIcon(instance uintptr) uintptr {
	// The build embeds the official dwrean.net icon as IDI_APPLICATION.
	return call(pLoadImage,instance,idiApplication,imageIcon,48,48,0)
}
