//go:build windows

package main

const idRecycle = 1003

var pSHEmptyRecycleBin = shell32.NewProc("SHEmptyRecycleBinW")

// This is deliberately a separate operation. The large one-click cleanup
// button NEVER empties the Recycle Bin. Native Windows shell API handles
// deletion on all drives; do not modify $Recycle.Bin manually.
func emptyRecycleBinWithConfirmation() {
	if active || mainWindow == 0 {
		return
	}
	const mbYesNo = 0x00000004
	const mbIconWarning = 0x00000030
	const mbDefaultNo = 0x00000100
	const idYes = 6
	message := "Θέλεις να αδειάσεις ΟΡΙΣΤΙΚΑ τον Κάδο Ανακύκλωσης;\n\n" +
		"Η ενέργεια αφορά τους Κάδους όλων των δίσκων του υπολογιστή.\n" +
		"Τα αρχεία που βρίσκονται εκεί δεν θα μπορούν να επανέλθουν από τον Κάδο.\n\n" +
		"Αν δεν είσαι βέβαιος, επίλεξε Όχι."
	answer := call(pMessageBox, mainWindow, ptr(message),
		ptr("Επιβεβαίωση • Άδειασμα Κάδου"),
		mbYesNo|mbIconWarning|mbDefaultNo)
	if answer != idYes {
		return
	}
	call(pEnableWindow, cleanButton, 0)
	call(pEnableWindow, recycleButton, 0)
	// The first confirmation was explicit; keep Windows' progress UI.
	// SHERB_NOCONFIRMATION prevents a duplicate Windows confirmation.
	const sherbNoConfirmation = 0x00000001
	hr := call(pSHEmptyRecycleBin, mainWindow, 0, sherbNoConfirmation)
	call(pEnableWindow, cleanButton, 1)
	call(pEnableWindow, recycleButton, 1)
	if hr == 0 {
		call(pMessageBox, mainWindow,
			ptr("Ο Κάδος Ανακύκλωσης άδειασε.\nΗ ενέργεια ήταν ανεξάρτητη από τον βασικό καθαρισμό."),
			ptr("dwrean Cleaner"), 0x00000040)
	} else {
		call(pMessageBox, mainWindow,
			ptr("Το άδειασμα του Κάδου δεν ολοκληρώθηκε ή ακυρώθηκε.\nΜπορείς να το κάνεις και από την επιφάνεια εργασίας των Windows."),
			ptr("dwrean Cleaner"), 0x00000030)
	}
}
