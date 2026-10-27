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

// A small, modern Win32 interface: no installer, webview, or third-party GUI runtime.
const (
	wsVisible     = 0x10000000
	wsChild       = 0x40000000
	wsCaption     = 0x00C00000
	wsSysMenu     = 0x00080000
	wsMinimizeBox = 0x00020000
	wsTabStop     = 0x00010000
	bsOwnerDraw   = 0x0000000B
	wmDestroy     = 0x0002
	wmPaint       = 0x000F
	wmCommand     = 0x0111
	wmDrawItem    = 0x002B
	wmSetFont     = 0x0030
	wmDone        = 0x8001
	wmClose       = 0x0010
	dtCenter      = 0x00000001
	dtVCenter     = 0x00000004
	dtSingleLine  = 0x00000020
	dtWordBreak   = 0x00000010
	dtLeft        = 0x00000000
	transparent   = 1
	swShow        = 5
	idClean       = 1001
	odsSelected   = 0x0001
	odsDisabled   = 0x0004
	odsFocus      = 0x0010
	odsHotlight   = 0x0040
	nullPen       = 8
	imageIcon     = 1
)

const (
	canvasW = 790
	canvasH = 575
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

	pRegisterClass   = user32.NewProc("RegisterClassExW")
	pCreateWindow    = user32.NewProc("CreateWindowExW")
	pDefWindow       = user32.NewProc("DefWindowProcW")
	pGetMessage      = user32.NewProc("GetMessageW")
	pTranslateMessage = user32.NewProc("TranslateMessage")
	pDispatchMessage = user32.NewProc("DispatchMessageW")
	pPostQuit        = user32.NewProc("PostQuitMessage")
	pSendMessage     = user32.NewProc("SendMessageW")
	pPostMessage     = user32.NewProc("PostMessageW")
	pShowWindow      = user32.NewProc("ShowWindow")
	pUpdateWindow    = user32.NewProc("UpdateWindow")
	pEnableWindow    = user32.NewProc("EnableWindow")
	pInvalidateRect  = user32.NewProc("InvalidateRect")
	pBeginPaint      = user32.NewProc("BeginPaint")
	pEndPaint        = user32.NewProc("EndPaint")
	pFillRect        = user32.NewProc("FillRect")
	pDrawText        = user32.NewProc("DrawTextW")
	pLoadCursor      = user32.NewProc("LoadCursorW")
	pDrawIconEx      = user32.NewProc("DrawIconEx")
	pSetDPIAware     = user32.NewProc("SetProcessDPIAware")
	pSetWindowText   = user32.NewProc("SetWindowTextW")
	pGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
	pCreateBrush     = gdi32.NewProc("CreateSolidBrush")
	pCreateFont      = gdi32.NewProc("CreateFontW")
	pSelectObject    = gdi32.NewProc("SelectObject")
	pDeleteObject    = gdi32.NewProc("DeleteObject")
	pGetStockObject  = gdi32.NewProc("GetStockObject")
	pRoundRect       = gdi32.NewProc("RoundRect")
	pSetBkMode       = gdi32.NewProc("SetBkMode")
	pSetTextColor    = gdi32.NewProc("SetTextColor")

	mainWindow uintptr
	cleanButton uintptr
	aboutButton uintptr
	backgroundBrush uintptr
	whiteBrush uintptr
	appIcon uintptr

	fontLogo uintptr
	fontTiny uintptr
	fontBody uintptr
	fontMedium uintptr
	fontTitle uintptr
	fontAction uintptr
	fontResult uintptr

	active bool
	finished bool
	statusTitle = "Έτοιμο για καθαρισμό"
	statusBody  = "Πάτησε το κουμπί για να αφαιρέσεις παλιά προσωρινά αρχεία."
	statusFoot  = "Windows   •   Browsers   •   GPU cache   •   Αναφορές σφαλμάτων"
	completed CleanResult
	completedMu sync.Mutex
)

var (
	cNavy = rgb(20, 36, 64)
	cMuted = rgb(91, 105, 128)
	cBlue = rgb(40, 104, 236)
	cBlueHover = rgb(27, 86, 204)
	cSoftBlue = rgb(236, 243, 255)
	cBorder = rgb(227, 234, 244)
	cCanvas = rgb(247, 249, 253)
	cSuccess = rgb(21, 133, 92)
	cWhite = rgb(255, 255, 255)
)

func rgb(r,g,b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }
func utf16(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }
func ptr(s string) uintptr { return uintptr(unsafe.Pointer(utf16(s))) }
func call(p *syscall.LazyProc, args ...uintptr) uintptr { r,_,_:=p.Call(args...); return r }
func rect(x,y,w,h int32) winRect { return winRect{x,y,x+w,y+h} }

func font(size, weight int) uintptr {
	return call(pCreateFont, uintptr(int32(-size)), 0,0,0,uintptr(weight),0,0,0,1,0,0,5,0,ptr("Segoe UI"))
}

func fillRect(hdc uintptr, r winRect, color uintptr) {
	b:=call(pCreateBrush,color)
	call(pFillRect,hdc,uintptr(unsafe.Pointer(&r)),b)
	call(pDeleteObject,b)
}

func roundRect(hdc uintptr, r winRect, color uintptr, radius int32) {
	b:=call(pCreateBrush,color)
	oldB:=call(pSelectObject,hdc,b)
	oldP:=call(pSelectObject,hdc,call(pGetStockObject,nullPen))
	call(pRoundRect,hdc,uintptr(r.Left),uintptr(r.Top),uintptr(r.Right),uintptr(r.Bottom),uintptr(radius),uintptr(radius))
	call(pSelectObject,hdc,oldP)
	call(pSelectObject,hdc,oldB)
	call(pDeleteObject,b)
}

func drawText(hdc uintptr, content string, r winRect, f, color uintptr, flags uintptr) {
	old:=call(pSelectObject,hdc,f)
	call(pSetBkMode,hdc,transparent)
	call(pSetTextColor,hdc,color)
	call(pDrawText,hdc,ptr(content),^uintptr(0),uintptr(unsafe.Pointer(&r)),flags)
	call(pSelectObject,hdc,old)
}

func button(parent uintptr, id int32, title string, r winRect) uintptr {
	w:=call(pCreateWindow,0,ptr("BUTTON"),ptr(title),
		wsVisible|wsChild|wsTabStop|bsOwnerDraw,
		uintptr(r.Left),uintptr(r.Top),uintptr(r.Right-r.Left),uintptr(r.Bottom-r.Top),
		parent,uintptr(id),0,0)
	call(pSendMessage,w,wmSetFont,fontBody,0)
	return w
}

func drawModernButton(item *drawItem) {
	r:=item.Rc
	pressed:=item.ItemState&odsSelected != 0
	hot:=item.ItemState&odsHotlight != 0
	disabled:=item.ItemState&odsDisabled != 0
	color:=cBlue
	label:=""
	textColor:=cWhite
	f:=fontBody

	switch item.CtlID {
	case idClean:
		label="ΚΑΘΑΡΙΣΜΟΣ ΤΩΡΑ"
		f=fontAction
		if active {label="ΓΙΝΕΤΑΙ ΚΑΘΑΡΙΣΜΟΣ..."}
		if finished && !active {label="ΚΑΘΑΡΙΣΜΟΣ ΞΑΝΑ"}
		if disabled {color=rgb(152,180,225)} else if pressed {color=rgb(21,73,180)} else if hot {color=cBlueHover}
	case idAbout:
		label="ⓘ  Περί"
		color=cSoftBlue
		textColor=cBlue
		if pressed {color=rgb(213,228,252)} else if hot {color=rgb(222,235,255)}
	case idSite:
		label="dwrean.net  ↗"
		color=cBlue
		if pressed {color=rgb(21,73,180)} else if hot {color=cBlueHover}
	case idRepo:
		label="GitHub  ↗"
		color=cSoftBlue
		textColor=cBlue
		if pressed {color=rgb(213,228,252)} else if hot {color=rgb(222,235,255)}
	case idDismiss:
		label="Κλείσιμο"
		color=rgb(240,244,250)
		textColor=cNavy
		if pressed {color=rgb(217,227,241)} else if hot {color=rgb(226,234,245)}
	}
	roundRect(item.Hdc,r,color,18)
	drawText(item.Hdc,label,r,f,textColor,dtCenter|dtVCenter|dtSingleLine)
	// Keep native button focus and keyboard navigation accessible.
	if item.ItemState&odsFocus != 0 {
		focus:=winRect{r.Left+4,r.Top+4,r.Right-4,r.Bottom-4}
		call(user32.NewProc("DrawFocusRect"),item.Hdc,uintptr(unsafe.Pointer(&focus)))
	}
}

func paintMain(hwnd uintptr) {
	var ps paintStruct
	hdc:=call(pBeginPaint,hwnd,uintptr(unsafe.Pointer(&ps)))
	if hdc==0 {return}

	fillRect(hdc,rect(0,0,canvasW,canvasH),cCanvas)
	fillRect(hdc,rect(0,0,canvasW,6),cBlue)
	// Brand header
	if appIcon!=0 {
		call(pDrawIconEx,hdc,40,29,appIcon,48,48,0,0,3)
	}
	drawText(hdc,"dwrean Cleaner",rect(104,29,435,38),fontLogo,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"PORTABLE  /  WINDOWS 10 & 11",rect(106,66,340,24),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	roundRect(hdc,rect(594,39,75,31),cWhite,13)
	drawText(hdc,"v"+appVersion,rect(594,39,75,31),fontTiny,cMuted,dtCenter|dtVCenter|dtSingleLine)

	// Intro section
	drawText(hdc,"Καθάρισε έξυπνα.",rect(42,121,705,46),fontTitle,cNavy,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"Λιγότερα περιττά αρχεία, περισσότερος ελεύθερος χώρος.",rect(44,172,691,30),fontBody,cMuted,dtLeft|dtVCenter|dtSingleLine)

	// Raised main cleanup surface
	roundRect(hdc,rect(40,226,710,225),rgb(231,237,246),32)
	roundRect(hdc,rect(38,222,714,225),cWhite,32)
	dotColor:=cSuccess
	if active {dotColor=cBlue}
	roundRect(hdc,rect(65,247,11,11),dotColor,11)
	header:="ΕΤΟΙΜΟ ΓΙΑ ΚΑΘΑΡΙΣΜΟ"
	if active {header="Ο ΚΑΘΑΡΙΣΜΟΣ ΕΙΝΑΙ ΣΕ ΕΞΕΛΙΞΗ"}
	if finished {header="Ο ΚΑΘΑΡΙΣΜΟΣ ΟΛΟΚΛΗΡΩΘΗΚΕ"}
	drawText(hdc,header,rect(85,239,600,28),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)

	statusColor:=cNavy
	if finished {statusColor=cSuccess}
	drawText(hdc,statusTitle,rect(67,353,649,29),fontMedium,statusColor,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,statusBody,rect(67,380,651,43),fontTiny,cMuted,dtLeft|dtWordBreak)

	// Subtle information banner
	roundRect(hdc,rect(38,462,714,59),cSoftBlue,22)
	roundRect(hdc,rect(57,479,20,20),rgb(202,224,255),20)
	drawText(hdc,"✓",rect(57,479,20,20),fontTiny,cBlue,dtCenter|dtVCenter|dtSingleLine)
	drawText(hdc,statusFoot,rect(90,469,637,45),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)

	drawText(hdc,"© 2026 dwrean.net    •    Δωρεάν και ανοιχτού κώδικα (MIT)",rect(43,532,600,25),fontTiny,cMuted,dtLeft|dtVCenter|dtSingleLine)
	drawText(hdc,"v"+appVersion,rect(676,532,71,25),fontTiny,cMuted,dtCenter|dtVCenter|dtSingleLine)
	call(pEndPaint,hwnd,uintptr(unsafe.Pointer(&ps)))
}

func repaintMain() {if mainWindow!=0 {call(pInvalidateRect,mainWindow,0,1)}}

func startClean() {
	if active {return}
	active=true
	finished=false
	statusTitle="Γίνεται καθαρισμός..."
	statusBody="Ελέγχουμε προσεκτικά τους επιτρεπόμενους φακέλους."
	statusFoot="Μην κλείσεις την εφαρμογή μέχρι να ολοκληρωθεί."
	call(pEnableWindow,cleanButton,0)
	repaintMain()
	go func(){
		running:=runningBrowsers()
		result:=cleanTargets(fullTargetsForLocal(os.Getenv("LOCALAPPDATA"),running),nowForCleanup())
		for _,name:=range []string{"chrome","msedge","brave","firefox"} {
			if running[name] {result.BrowsersSkipped++}
		}
		completedMu.Lock()
		completed=result
		completedMu.Unlock()
		call(pPostMessage,mainWindow,wmDone,0,0)
	}()
}

func finishClean() {
	completedMu.Lock()
	result:=completed
	completedMu.Unlock()
	active=false
	finished=true
	if result.Locations==0 {
		statusTitle="Δεν βρέθηκαν διαθέσιμοι φάκελοι."
		statusBody="Δεν πραγματοποιήθηκε καμία διαγραφή."
	}else if result.Deleted==0 {
		statusTitle="Δεν βρέθηκαν παλιά προσωρινά αρχεία."
		statusBody="Δεν διαγράφηκε κανένα αρχείο σε αυτόν τον έλεγχο."
	}else{
		statusTitle=fmt.Sprintf("Ελευθερώθηκαν %s",formatBytes(result.Bytes))
		statusBody=fmt.Sprintf("Διαγράφηκαν %d παλιά προσωρινά αρχεία.",result.Deleted)
		if result.Skipped>0 {statusBody+=" Κάποια προστατευμένα αρχεία παραλείφθηκαν."}
	}
	statusFoot="Τα προσωπικά σου αρχεία και οι ρυθμίσεις παραμένουν ανέπαφα."
	if result.BrowsersSkipped>0 {
		statusFoot="Για επιπλέον browser cache, κλείσε τους browsers και ξαναδοκίμασε."
	}
	call(pEnableWindow,cleanButton,1)
	repaintMain()
}

func mainWndProc(hwnd uintptr,msg uint32,wparam,lparam uintptr) uintptr {
	switch msg{
	case wmPaint:
		paintMain(hwnd)
		return 0
	case wmDrawItem:
		if lparam!=0 {
			item:=(*drawItem)(unsafe.Pointer(lparam))
			if item.CtlID==idClean || item.CtlID==idAbout {
				drawModernButton(item)
				return 1
			}
		}
	case wmCommand:
		switch uint32(wparam&0xffff) {
		case idClean:
			startClean()
			return 0
		case idAbout:
			showAbout()
			return 0
		}
	case wmDone:
		finishClean()
		return 0
	case wmDestroy:
		mainWindow=0
		for _,h:=range []uintptr{fontLogo,fontTiny,fontBody,fontMedium,fontTitle,fontAction,fontResult,
			backgroundBrush,whiteBrush,aboutTitleFont} {
			if h!=0 {call(pDeleteObject,h)}
		}
		call(pPostQuit,0)
		return 0
	}
	return call(pDefWindow,hwnd,uintptr(msg),wparam,lparam)
}

func main() {
	runtime.LockOSThread()
	// Windows scales a single compact UI uniformly for standard DPI settings.
	// Do not force DPI-awareness without scaling every custom-painted element.
	backgroundBrush=call(pCreateBrush,cCanvas)
	whiteBrush=call(pCreateBrush,cWhite)
	fontLogo=font(24,700)
	fontTiny=font(13,500)
	fontBody=font(16,400)
	fontMedium=font(18,600)
	fontTitle=font(34,700)
	fontAction=font(20,700)
	fontResult=font(23,700)
	instance:=call(pGetModuleHandle,0)
	appIcon=loadAppIcon(instance)
	className:=utf16("DwreanCleanerModernV11")
	class:=wndClass{
		CbSize:uint32(unsafe.Sizeof(wndClass{})),
		WndProc:syscall.NewCallback(mainWndProc),
		Instance:instance,Icon:appIcon,IconSmall:appIcon,
		Cursor:call(pLoadCursor,0,32512),
		Background:backgroundBrush,ClassName:className,
	}
	if call(pRegisterClass,uintptr(unsafe.Pointer(&class)))==0 {return}
	// Outer size includes caption/frame; 810 x 625 fits common laptop screens.
	mainWindow=call(pCreateWindow,0,uintptr(unsafe.Pointer(className)),
		ptr("dwrean Cleaner Portable  •  v"+appVersion),
		wsCaption|wsSysMenu|wsMinimizeBox,0x80000000,0x80000000,790,615,0,0,instance,0)
	if mainWindow==0 {return}
	aboutButton=button(mainWindow,idAbout,"Περί",rect(682,38,69,33))
	cleanButton=button(mainWindow,idClean,"Καθαρισμός",rect(65,275,661,66))
	call(pShowWindow,mainWindow,swShow)
	call(pUpdateWindow,mainWindow)
	var message winMessage
	for {
		result:=int32(call(pGetMessage,uintptr(unsafe.Pointer(&message)),0,0,0))
		if result<=0 {break}
		call(pTranslateMessage,uintptr(unsafe.Pointer(&message)))
		call(pDispatchMessage,uintptr(unsafe.Pointer(&message)))
	}
}
