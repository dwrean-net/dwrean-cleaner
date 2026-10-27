<div align="center">

<a href="https://www.dwrean.net/">
  <img src="https://raw.githubusercontent.com/dwrean-net/dwrean-radio/main/assets/dwrean-logo.png" alt="Λογότυπο dwrean.net" width="92">
</a>

# dwrean Cleaner Portable

### Καθάρισε τα Windows με ένα κλικ.

**Δωρεάν · Portable · Ελληνικό περιβάλλον · Ανοιχτός κώδικας**

[![Stable release](https://img.shields.io/github/v/release/dwrean-net/dwrean-cleaner?label=Stable%20version&style=flat-square&color=2775ea)](https://github.com/dwrean-net/dwrean-cleaner/releases/latest)
[![Build](https://github.com/dwrean-net/dwrean-cleaner/actions/workflows/build.yml/badge.svg)](https://github.com/dwrean-net/dwrean-cleaner/actions/workflows/build.yml)
[![MIT](https://img.shields.io/badge/License-MIT-15855c?style=flat-square)](LICENSE)
[![Windows 10/11](https://img.shields.io/badge/Windows-10%20%2F%2011-435d7d?style=flat-square)](#απαιτήσεις)

### [⬇ Κατέβασε την τελευταία σταθερή έκδοση](https://github.com/dwrean-net/dwrean-cleaner/releases/latest)

[Ιστορικό αλλαγών](CHANGELOG.md) · [Αναφορά προβλήματος](https://github.com/dwrean-net/dwrean-cleaner/issues/new) · [dwrean.net](https://www.dwrean.net/)

</div>

---

## Τι είναι;

Το **dwrean Cleaner Portable** είναι ένα μικρό, δωρεάν πρόγραμμα για **Windows 10/11**. Εντοπίζει και αφαιρεί **παλιά προσωρινά αρχεία** από συγκεκριμένους επιτρεπόμενους φακέλους, χωρίς περίπλοκες ρυθμίσεις, εγκατάσταση ή λογαριασμό. Απευθύνεται σε αρχάριους χρήστες.

| Χαρακτηριστικό | Τι προσφέρει |
|---|---|
| **Ένα κουμπί** | Πάτησε «ΚΑΘΑΡΙΣΜΟΣ ΤΩΡΑ». |
| **Portable** | Αυτόνομο EXE, κατάλληλο και για USB. |
| **Συντηρητικός καθαρισμός** | Διαγράφει μόνο παλιά αρχεία από καθορισμένους φακέλους. |
| **Εμφάνιση αποτελέσματος** | Αριθμός αρχείων και μέγεθος χώρου που ελευθερώθηκε. |
| **Μοντέρνο περιβάλλον** | Καθαρή σχεδίαση, επίσημο εικονίδιο και ανανεωμένο «Περί». |
| **Ανοιχτός κώδικας** | Go / Win32, άδεια MIT. |
| **Χωρίς τηλεμετρία** | Δεν συλλέγει πληροφορίες ούτε συνδέεται στο διαδίκτυο κατά τη χρήση. |

## Λήψη και χρήση

1. Άνοιξε την [**επίσημη σελίδα λήψεων**](https://github.com/dwrean-net/dwrean-cleaner/releases/latest).
2. Στο **Assets**, κατέβασε το **dwrean-cleaner-portable.exe** ή το ZIP.
3. Άνοιξε την εφαρμογή. Δεν απαιτείται εγκατάσταση ή εκτέλεση ως διαχειριστής.
4. Πάτησε το μεγάλο κουμπί **«ΚΑΘΑΡΙΣΜΟΣ ΤΩΡΑ»**.
5. Δες το αποτέλεσμα στην κάρτα καθαρισμού.

> Για πιο ολοκληρωμένο καθαρισμό των caches, **κλείσε πρώτα τους browsers**. Η εφαρμογή παραλείπει τις caches του Chrome, Edge, Brave ή Firefox αν το αντίστοιχο πρόγραμμα είναι ανοιχτό.

## Τι ακριβώς καθαρίζει;

Οι παρακάτω φάκελοι είναι κάτω από το προσωπικό **AppData\Local**, όχι από ολόκληρο τον δίσκο. Διαγράφονται μόνο *κανονικά αρχεία* με τελευταία τροποποίηση παλαιότερη από το αντίστοιχο όριο.

| Κατηγορία | Πόσο παλιά πρέπει να είναι |
|---|---:|
| Προσωρινά αρχεία χρήστη (Temp) | 14 ημέρες |
| DirectX shader cache | 14 ημέρες |
| Crash dumps εφαρμογών | 30 ημέρες |
| Αρχειοθετημένες αναφορές Windows Error Reporting | 30 ημέρες |
| Αναφορές Windows Error Reporting σε αναμονή | 45 ημέρες |
| Προσωρινές αναφορές Windows Error Reporting | 14 ημέρες |
| Επιλεγμένες caches NVIDIA και AMD | 30 ημέρες |
| Επιλεγμένη cache Chrome, Edge και Brave | 30 ημέρες |
| Επιλεγμένη cache Firefox | 30 ημέρες |

Τα αρχεία που χρησιμοποιούνται, τα μη προσβάσιμα αρχεία και οι συμβολικοί σύνδεσμοι / Windows junctions παραλείπονται. Οι caches αναδημιουργούνται όταν χρειαστούν· συνεπώς η εξοικονόμηση μπορεί να είναι προσωρινή.

### Τι ΔΕΝ διαγράφει

Δεν αφαιρεί **Έγγραφα, Εικόνες, Βίντεο, Λήψεις, Επιφάνεια εργασίας, Κάδο Ανακύκλωσης, cookies, ιστορικό, αποθηκευμένους κωδικούς, σελιδοδείκτες, sessions, Registry, εγκατεστημένα προγράμματα ή αρχεία Windows Update**.

> **Προσοχή:** Οι διαγραφές γίνονται **οριστικά**, χωρίς μεταφορά στον Κάδο. Παλιά αρχεία σφαλμάτων μπορεί να χρειάζονται για τεχνική διάγνωση. Δεν υπόσχεται επιτάχυνση ή συγκεκριμένα GB, ούτε αντικαθιστά τη λειτουργία **Storage Sense** των Windows.

## Απαιτήσεις

- **Windows 10 ή Windows 11**, 64-bit (x64).
- Δεν απαιτούνται εγκατάσταση, .NET runtime ή δικαιώματα διαχειριστή.
- Δεν χρειάζεται σύνδεση στο διαδίκτυο για τον καθαρισμό.

**Windows SmartScreen:** Το εκτελέσιμο **δεν έχει ψηφιακή υπογραφή εκδότη**. Είναι πιθανή προειδοποίηση για άγνωστη εφαρμογή. Προτίμησε πάντα τις [επίσημες λήψεις](https://github.com/dwrean-net/dwrean-cleaner/releases/latest) και, αν θέλεις, σύγκρινε το SHA-256 με το `SHA256SUMS.txt` του Release.

## Συχνές ερωτήσεις

<details>
<summary><strong>Γιατί δεν ελευθέρωσε πολύ χώρο;</strong></summary>

Η εφαρμογή καθαρίζει μόνο συγκεκριμένες παλιές caches και προσωρινά. Ίσως δεν υπάρχουν αρκετά παλιά αρχεία, κάποιο πρόγραμμα είναι ανοιχτό ή η cache έχει ήδη καθαριστεί.

</details>

<details>
<summary><strong>Γιατί δεν καθάρισε τον browser;</strong></summary>

Αν ο αντίστοιχος browser είναι ενεργός, ολόκληρος ο καθαρισμός της cache του παραλείπεται. Κλείσ' τον και δοκίμασε ξανά.

</details>

<details>
<summary><strong>Διαγράφει λήψεις, κωδικούς ή cookies;</strong></summary>

Όχι. Αυτά βρίσκονται εκτός των συγκεκριμένων στόχων καθαρισμού.

</details>

<details>
<summary><strong>Θα γίνει ταχύτερος ο υπολογιστής μου;</strong></summary>

Όχι απαραίτητα. Το όφελος αφορά κυρίως τη διαχείριση του αποθηκευτικού χώρου. Μερικές caches μπορεί να χρειαστούν επαναδημιουργία.

</details>

<details>
<summary><strong>Είναι απολύτως ακίνδυνος ο αυτόματος καθαρισμός;</strong></summary>

Καμία μόνιμη διαγραφή αρχείων δεν είναι εντελώς χωρίς κίνδυνο. Το εργαλείο περιορίζεται σε επιτρεπόμενους προσωρινούς φακέλους, αποφεύγει links και πρόσφατα/κλειδωμένα αρχεία. Για σημαντικά δεδομένα διατήρησε αντίγραφα ασφαλείας.

</details>

## Διαφάνεια και τεχνικές πληροφορίες

- **Δημιουργός:** Κυριάκος Οικονομίδης — [dwrean.net](https://www.dwrean.net/)
- **Άδεια:** [MIT](LICENSE) · Δωρεάν και ανοικτός κώδικας.
- **Τεχνολογία:** Go και Windows API (Win32), χωρίς ενσωματωμένο browser ή εξωτερικό GUI runtime.
- **Ιδιωτικότητα:** Δεν συλλέγονται δεδομένα, δεν υπάρχουν διαφημίσεις, παρακολούθηση ή αυτόματα background tasks.
- **Έλεγχοι:** Το [GitHub Actions](.github/workflows/build.yml) εκτελεί αυτοματοποιημένες δοκιμές σε Linux και Windows και παράγει EXE / ZIP / SHA-256. Οι έλεγχοι αυτοί δεν εξασφαλίζουν δοκιμή σε κάθε πραγματικό PC.
- **Κυκλοφορίες:** Στο [Releases](https://github.com/dwrean-net/dwrean-cleaner/releases) δημοσιεύονται οι κανονικές σταθερές εκδόσεις· οι παλιές beta διατηρούνται μόνο για ιστορικό.

### Build από τον πηγαίο κώδικα

Χρειάζεται Go 1.22+ σε Windows. Για βασικό build (χωρίς το ενσωματωμένο εικονίδιο / version resource):

~~~powershell
go test ./...
go build -trimpath -ldflags="-s -w -H windowsgui" -o dwrean-cleaner-portable.exe .
~~~

Για την **κανονική branded έκδοση** χρησιμοποιείται το [workflow του GitHub](.github/workflows/build.yml), που ενσωματώνει το επίσημο λογότυπο και τα Windows file properties.

## Υποστήριξη και αναφορές σφαλμάτων

Έχεις πρόταση ή παρατήρησες πρόβλημα; [Δημιούργησε ένα GitHub Issue](https://github.com/dwrean-net/dwrean-cleaner/issues/new). Γράψε την έκδοση του προγράμματος, την έκδοση των Windows και τα βήματα που προκάλεσαν το πρόβλημα — χωρίς προσωπικά αρχεία ή ευαίσθητες πληροφορίες.

Δες επίσης το [CHANGELOG](CHANGELOG.md) και τους [όρους χρήσης MIT](LICENSE).

---

<div align="center">

**Ένα κλικ. Λιγότερα περιττά αρχεία.**

Δημιουργήθηκε από τον **Κυριάκο Οικονομίδη** για το [dwrean.net](https://www.dwrean.net/).

</div>
