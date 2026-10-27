//go:build windows

package main

import (
    "syscall"
    "unsafe"
)

const (
    appVersion = "1.0.0"
    appSiteURL = "https://www.dwrean.net/"
    appRepoURL = "https://github.com/dwrean-net/dwrean-cleaner"
    idAbout = 1002
    idSite = 2001
    idRepo = 2002
    idDismiss = 2003
    wmSetIcon = 0x0080
    imageIcon = 1
    idiApplication = 32512
)

var (
    shell32 = syscall.NewLazyDLL("shell32.dll")
    pShellExecute = shell32.NewProc("ShellExecuteW")
    pMessageBox = user32.NewProc("MessageBoxW")
    pLoadImage = user32.NewProc("LoadImageW")
    hwndAbout uintptr
    aboutClassRegistered bool
    aboutClassName = utf16("DwreanCleanerAboutWindowV1")
    aboutFont uintptr
    aboutSmallFont uintptr
    aboutLogoIcon uintptr
)

func openOfficialURL(url string) {
    // Only fixed links controlled by the project's publisher can be opened.
    if url != appSiteURL && url != appRepoURL { return }
    ret := call(pShellExecute, hwndMain, ptr("open"), ptr(url), 0, 0, 1)
    if ret <= 32 {
        call(pMessageBox, hwndMain, ptr("Δεν ήταν δυνατή η εκκίνηση του browser."), ptr("dwrean Cleaner"), 0x30)
    }
}

func buttonIn(parent uintptr, text string, id, x, y, width, height int32) uintptr {
    hwnd := call(pCreateWindow, 0, ptr("BUTTON"), ptr(text), wsChild|wsVisible|0x00010000,
        uintptr(x),uintptr(y),uintptr(width),uintptr(height), parent,uintptr(id),0,0)
    setFont(hwnd, fontDescription)
    return hwnd
}

func labelIn(parent uintptr, text string, x, y, w, h int32, font uintptr) uintptr {
    hwnd := call(pCreateWindow,0,ptr("STATIC"),ptr(text), wsChild|wsVisible,
        uintptr(x),uintptr(y),uintptr(w),uintptr(h),parent,0,0,0)
    setFont(hwnd,font)
    return hwnd
}

func aboutWndProc(hwnd uintptr, msg uint32, wparam,lparam uintptr) uintptr {
    switch msg {
    case wmCtlColorStatic:
        call(pSetBkMode,wparam,transparent)
        call(pSetTextColor,wparam,rgb(26,45,72))
        return brushWhite
    case wmCommand:
        switch int(wparam & 0xffff) {
        case idSite:
            openOfficialURL(appSiteURL)
        case idRepo:
            openOfficialURL(appRepoURL)
        case idDismiss:
            call(user32.NewProc("DestroyWindow"), hwnd)
        }
        return 0
    case wmClose:
        call(user32.NewProc("DestroyWindow"), hwnd)
        return 0
    case wmDestroy:
        hwndAbout=0
        return 0
    }
    return call(pDefWindow,hwnd,uintptr(msg),wparam,lparam)
}

func showAbout() {
    if hwndAbout!=0 {
        call(user32.NewProc("SetForegroundWindow"),hwndAbout)
        return
    }
    instance:=call(pGetModuleHandle,0)
    if !aboutClassRegistered {
        class:=wndClass{CbSize:uint32(unsafe.Sizeof(wndClass{})),WndProc:syscall.NewCallback(aboutWndProc),Instance:instance,
            Cursor:call(pLoadCursor,0,32512),Background:brushWhite,ClassName:aboutClassName,
            Icon:aboutLogoIcon,IconSmall:aboutLogoIcon}
        if call(pRegisterClass,uintptr(unsafe.Pointer(&class)))==0 {return}
        aboutClassRegistered=true
    }
    hwndAbout=call(pCreateWindow,0,uintptr(unsafe.Pointer(aboutClassName)),ptr("Περί του dwrean Cleaner"),
        wsCaption|wsSysMenu,0x80000000,0x80000000,500,367,hwndMain,0,instance,0)
    if hwndAbout==0 {return}
    if aboutLogoIcon!=0 {call(pSendMessage,hwndAbout,wmSetIcon,0,aboutLogoIcon)}
    if aboutFont==0 {aboutFont=createFont(21,700,"Segoe UI")}
    if aboutSmallFont==0 {aboutSmallFont=createFont(15,400,"Segoe UI")}
    labelIn(hwndAbout,"dwrean Cleaner Portable",25,25,440,33,aboutFont)
    labelIn(hwndAbout,"Έκδοση "+appVersion+"  •  Windows 10 / 11",26,64,445,28,aboutSmallFont)
    labelIn(hwndAbout,"Δημιουργός: Κυριάκος Οικονομίδης",26,103,445,26,aboutSmallFont)
    labelIn(hwndAbout,"Μια δωρεάν εφαρμογή από το dwrean.net",26,133,445,26,aboutSmallFont)
    labelIn(hwndAbout,"Ανοιχτός κώδικας (MIT) • Χωρίς εγκατάσταση",26,166,445,26,aboutSmallFont)
    buttonIn(hwndAbout,"Επίσκεψη στο dwrean.net",idSite,25,209,211,39)
    buttonIn(hwndAbout,"Κώδικας στο GitHub",idRepo,249,209,214,39)
    buttonIn(hwndAbout,"Κλείσιμο",idDismiss,344,272,119,37)
    call(pShowWindow,hwndAbout,swShow)
    call(pUpdateWindow,hwndAbout)
}

func loadAppIcon(instance uintptr) uintptr {
    // GoVersionInfo embeds the dwrean logo as the IDI_APPLICATION icon resource.
    icon:=call(pLoadImage,instance,idiApplication,imageIcon,48,48,0)
    if icon==0 {return 0}
    aboutLogoIcon=icon
    return icon
}
