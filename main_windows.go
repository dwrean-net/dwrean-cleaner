//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// A dependency-free, native Win32 GUI: this builds into a single portable EXE.
const (
	wsVisible        = 0x10000000
	wsChild          = 0x40000000
	wsCaption        = 0x00C00000
	wsSysMenu        = 0x00080000
	wsMinimizeBox    = 0x00020000
	bsOwnerDraw      = 0x0000000B
	wmDestroy        = 0x0002
	wmPaint          = 0x000F
	wmCommand        = 0x0111
	wmDrawItem       = 0x002B
	wmSetFont        = 0x0030
	wmCtlColorStatic = 0x0138
	wmDone           = 0x8001
	wmClose          = 0x0010
	dtCenter         = 0x00000001
	dtVCenter        = 0x00000004
	dtSingleLine     = 0x00000020
	transparent      = 1
	swShow           = 5
	idClean          = 1001
	odsDisabled      = 0x0004
	odsSelected      = 0x0001
	nullPen          = 8
)

type winRect struct{ Left, Top, Right, Bottom int32 }
type winPoint struct{ X, Y int32 }
type winMessage struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       winPoint
	LPrivate uint32
}
type wndClass struct {
	CbSize     uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSmall  uintptr
}
type paintStruct struct {
	Hdc       uintptr
	Erase     int32
	Paint     winRect
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}
type drawItem struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   uintptr
	Hdc       uintptr
	Rc         winRect
	ItemData   uintptr
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pRegisterClass      = user32.NewProc("RegisterClassExW")
	pCreateWindow       = user32.NewProc("CreateWindowExW")
	pDefWindow          = user32.NewProc("DefWindowProcW")
	pGetMessage         = user32.NewProc("GetMessageW")
	pTranslateMessage   = user32.NewProc("TranslateMessage")
	pDispatchMessage    = user32.NewProc("DispatchMessageW")
	pPostQuit           = user32.NewProc("PostQuitMessage")
	pSendMessage        = user32.NewProc("SendMessageW")
	pPostMessage        = user32.NewProc("PostMessageW")
	pShowWindow         = user32.NewProc("ShowWindow")
	pUpdateWindow       = user32.NewProc("UpdateWindow")
	pSetWindowText      = user32.NewProc("SetWindowTextW")
	pEnableWindow       = user32.NewProc("EnableWindow")
	pInvalidateRect     = user32.NewProc("InvalidateRect")
	pBeginPaint         = user32.NewProc("BeginPaint")
	pEndPaint           = user32.NewProc("EndPaint")
	pFillRect           = user32.NewProc("FillRect")
	pDrawText           = user32.NewProc("DrawTextW")
	pGetSysColor        = user32.NewProc("GetSysColor")
	pLoadCursor         = user32.NewProc("LoadCursorW")
	pSetProcessDPIAware = user32.NewProc("SetProcessDPIAware")

	pGetModuleHandle  = kernel32.NewProc("GetModuleHandleW")
	pCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	pCreateFont       = gdi32.NewProc("CreateFontW")
	pSelectObject     = gdi32.NewProc("SelectObject")
	pDeleteObject     = gdi32.NewProc("DeleteObject")
	pGetStockObject   = gdi32.NewProc("GetStockObject")
	pRoundRect        = gdi32.NewProc("RoundRect")
	pSetBkMode        = gdi32.NewProc("SetBkMode")
	pSetTextColor     = gdi32.NewProc("SetTextColor")

	hwndMain        uintptr
	hwndButton      uintptr
	hwndStatus      uintptr
	hwndNote        uintptr
	hwndHeader      uintptr
	hwndKicker      uintptr
	hwndTitle       uintptr
	hwndDescription uintptr
	hwndFooter      uintptr

	brushBackground uintptr
	brushWhite      uintptr
	fontBrand       uintptr
	fontKicker      uintptr
	fontTitle       uintptr
	fontDescription uintptr
	fontButton      uintptr
	fontStatus      uintptr
	fontFooter      uintptr

	active      bool
	finished    bool
	completed   CleanResult
	completedMu sync.Mutex
)

func rgb(r, g, b byte) uintptr                             { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }
func utf16(s string) *uint16                               { return syscall.StringToUTF16Ptr(s) }
func call(proc *syscall.LazyProc, args ...uintptr) uintptr { r, _, _ := proc.Call(args...); return r }
func ptr(s string) uintptr                                 { return uintptr(unsafe.Pointer(utf16(s))) }

func createFont(height, weight int, face string) uintptr {
	return call(pCreateFont, uintptr(int32(-height)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, ptr(face))
}
func setFont(hwnd, font uintptr) { call(pSendMessage, hwnd, wmSetFont, font, 1) }
func label(content string, x, y, w, h int32, font uintptr) uintptr {
	hwnd := call(pCreateWindow, 0, ptr("STATIC"), ptr(content), wsChild|wsVisible,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h), hwndMain, 0, 0, 0)
	setFont(hwnd, font)
	return hwnd
}
func repaint()                       { call(pInvalidateRect, hwndMain, 0, 1) }
func setText(hwnd uintptr, s string) { call(pSetWindowText, hwnd, ptr(s)) }

func roundedFill(hdc uintptr, rect winRect, fill uintptr, radius int32) {
	brush := call(pCreateSolidBrush, fill)
	oldBrush := call(pSelectObject, hdc, brush)
	pen := call(pGetStockObject, nullPen)
	oldPen := call(pSelectObject, hdc, pen)
	call(pRoundRect, hdc, uintptr(rect.Left), uintptr(rect.Top), uintptr(rect.Right), uintptr(rect.Bottom), uintptr(radius), uintptr(radius))
	call(pSelectObject, hdc, oldPen)
	call(pSelectObject, hdc, oldBrush)
	call(pDeleteObject, brush)
}

func drawBackground(hwnd uintptr) {
	var ps paintStruct
	hdc := call(pBeginPaint, hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	var accent = winRect{0, 0, 680, 7}
	blueBrush := call(pCreateSolidBrush, rgb(31, 110, 245))
	call(pFillRect, hdc, uintptr(unsafe.Pointer(&accent)), blueBrush)
	call(pDeleteObject, blueBrush)
	roundedFill(hdc, winRect{33, 231, 647, 368}, rgb(255, 255, 255), 24)
	call(pEndPaint, hwnd, uintptr(unsafe.Pointer(&ps)))
}

func drawButton(item *drawItem) {
	color := rgb(31, 110, 245)
	if active || item.ItemState&odsDisabled != 0 {
		color = rgb(149, 176, 217)
	}
	if item.ItemState&odsSelected != 0 && !active {
		color = rgb(24, 87, 196)
	}
	roundedFill(item.Hdc, item.Rc, color, 24)
	call(pSetBkMode, item.Hdc, transparent)
	call(pSetTextColor, item.Hdc, rgb(255, 255, 255))
	oldFont := call(pSelectObject, item.Hdc, fontButton)
	text := "ΚΑΘΑΡΙΣΜΟΣ ΤΩΡΑ"
	if active {
		text = "ΓΙΝΕΤΑΙ ΚΑΘΑΡΙΣΜΟΣ..."
	} else if finished {
		text = "ΚΑΘΑΡΙΣΜΟΣ ΞΑΝΑ"
	}
	r := item.Rc
	call(pDrawText, item.Hdc, ptr(text), ^uintptr(0), uintptr(unsafe.Pointer(&r)), dtCenter|dtVCenter|dtSingleLine)
	call(pSelectObject, item.Hdc, oldFont)
}

func startClean() {
	if active {
		return
	}
	active = true
	finished = false
	setText(hwndStatus, "Ελέγχονται Windows, προσωρινά, browser και GPU cache...")
	setText(hwndNote, "Δεν αφαιρούνται προσωπικά αρχεία ή αρχεία που χρησιμοποιούνται.")
	call(pEnableWindow, hwndButton, 0)
	repaint()
	go func() {
		result := cleanTargets(fullTargetsForLocal(os.Getenv("LOCALAPPDATA"), runningBrowsers()), nowForCleanup())
		completedMu.Lock()
		completed = result
		completedMu.Unlock()
		call(pPostMessage, hwndMain, wmDone, 0, 0)
	}()
}

func onDone() {
	completedMu.Lock()
	result := completed
	completedMu.Unlock()
	active = false
	finished = true
	if result.Locations == 0 {
		setText(hwndStatus, "Δεν βρέθηκαν προσβάσιμοι φάκελοι καθαρισμού.")
		setText(hwndNote, "Δεν πραγματοποιήθηκε καμία διαγραφή.")
	} else if result.Deleted == 0 {
		setText(hwndStatus, "Ο υπολογιστής είναι ήδη καθαρός από παλιά προσωρινά αρχεία.")
		setText(hwndNote, "Δεν διαγράφηκε κανένα αρχείο.")
	} else {
		setText(hwndStatus, fmt.Sprintf("Αφαιρέθηκαν %d αρχεία · %s", result.Deleted, formatBytes(result.Bytes)))
		if result.Skipped > 0 {
			setText(hwndNote, "Ορισμένα προστατευμένα ή ενεργά αρχεία παραλείφθηκαν.")
		} else {
			setText(hwndNote, "Ο καθαρισμός ολοκληρώθηκε επιτυχώς.")
		}
	}
	call(pEnableWindow, hwndButton, 1)
	repaint()
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmPaint:
		drawBackground(hwnd)
		return 0
	case wmCtlColorStatic:
		call(pSetBkMode, wParam, transparent)
		switch lParam {
		case hwndHeader, hwndTitle:
			call(pSetTextColor, wParam, rgb(19, 39, 67))
		case hwndKicker:
			call(pSetTextColor, wParam, rgb(31, 110, 245))
		case hwndStatus:
			if finished {
				call(pSetTextColor, wParam, rgb(23, 120, 87))
			} else {
				call(pSetTextColor, wParam, rgb(27, 53, 81))
			}
			return brushWhite
		case hwndNote:
			call(pSetTextColor, wParam, rgb(109, 120, 139))
			return brushWhite
		default:
			call(pSetTextColor, wParam, rgb(109, 120, 139))
		}
		return brushBackground
	case wmDrawItem:
		item := (*drawItem)(unsafe.Pointer(lParam))
		if item.CtlID == idClean {
			drawButton(item)
			return 1
		}
	case wmCommand:
		if wParam&0xffff == idClean {
			startClean()
			return 0
		}
	case wmDone:
		onDone()
		return 0
	case wmClose:
		// Default close behavior, even while scanning; no new cleanup starts.
	case wmDestroy:
		for _, h := range []uintptr{brushBackground, brushWhite, fontBrand, fontKicker, fontTitle, fontDescription, fontButton, fontStatus, fontFooter} {
			if h != 0 {
				call(pDeleteObject, h)
			}
		}
		call(pPostQuit, 0)
		return 0
	}
	return call(pDefWindow, hwnd, uintptr(msg), wParam, lParam)
}

func buildUI() {
	fontBrand = createFont(25, 700, "Segoe UI")
	fontKicker = createFont(14, 600, "Segoe UI")
	fontTitle = createFont(29, 700, "Segoe UI")
	fontDescription = createFont(16, 400, "Segoe UI")
	fontButton = createFont(19, 700, "Segoe UI")
	fontStatus = createFont(15, 600, "Segoe UI")
	fontFooter = createFont(13, 400, "Segoe UI")

	hwndHeader = label("dwrean Cleaner", 44, 32, 590, 36, fontBrand)
	hwndKicker = label("PORTABLE  •  WINDOWS 10 / 11", 46, 77, 570, 23, fontKicker)
	hwndTitle = label("Ένας καθαρισμός. Ένα κλικ.", 44, 121, 605, 43, fontTitle)
	hwndDescription = label("Αφαίρεσε παλιές caches και προσωρινά αρχεία, με ένα πάτημα.", 45, 174, 596, 27, fontDescription)
	hwndButton = call(pCreateWindow, 0, ptr("BUTTON"), ptr("ΚΑΘΑΡΙΣΜΟΣ ΤΩΡΑ"), wsChild|wsVisible|bsOwnerDraw,
		51, 246, 576, 65, hwndMain, idClean, 0, 0)
	setFont(hwndButton, fontButton)
	hwndStatus = label("Πάτησε το κουμπί για να ξεκινήσεις.", 52, 323, 568, 24, fontStatus)
	hwndNote = label("Windows  •  Browser cache  •  GPU cache  •  Παλιά σφάλματα", 52, 348, 568, 20, fontFooter)
	hwndFooter = label("Οριστική διαγραφή προσωρινών · Λήψεις, Έγγραφα και Κάδος ανέπαφα.", 44, 390, 595, 24, fontFooter)
}

func main() {
	runtime.LockOSThread()
	call(pSetProcessDPIAware)
	brushBackground = call(pCreateSolidBrush, rgb(247, 249, 253))
	brushWhite = call(pCreateSolidBrush, rgb(255, 255, 255))
	instance := call(pGetModuleHandle, 0)
	classname := utf16("DwreanCleanerPortableWindowV1")
	cursor := call(pLoadCursor, 0, 32512) // IDC_ARROW
	wc := wndClass{CbSize: uint32(unsafe.Sizeof(wndClass{})), WndProc: syscall.NewCallback(wndProc), Instance: instance,
		Cursor: cursor, Background: brushBackground, ClassName: classname}
	if call(pRegisterClass, uintptr(unsafe.Pointer(&wc))) == 0 {
		return
	}
	hwndMain = call(pCreateWindow, 0, uintptr(unsafe.Pointer(classname)), ptr("dwrean Cleaner — Portable"),
		wsCaption|wsSysMenu|wsMinimizeBox, 0x80000000, 0x80000000, 680, 460, 0, 0, instance, 0)
	if hwndMain == 0 {
		return
	}
	buildUI()
	call(pShowWindow, hwndMain, swShow)
	call(pUpdateWindow, hwndMain)
	var message winMessage
	for {
		n := int32(call(pGetMessage, uintptr(unsafe.Pointer(&message)), 0, 0, 0))
		if n <= 0 {
			break
		}
		call(pTranslateMessage, uintptr(unsafe.Pointer(&message)))
		call(pDispatchMessage, uintptr(unsafe.Pointer(&message)))
	}
}
