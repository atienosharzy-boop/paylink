# PayLink: M-Pesa School Fee Tracker & Payment Links

**PayLink** is an M-Pesa fee collection and balance tracking platform built specifically for **Kenyan primary/secondary schools, private academies, and tutors**.

Instead of sharing a generic Paybill number and manually cross-referencing hundreds of bank slips and WhatsApp M-Pesa screenshots, PayLink gives **each student their own persistent payment link**.

Parents click the link, choose whether to pay in full or make an installment, approve the prompt on their phone, and an official digital fee receipt is issued immediately.

---

## 🌟 Key Features

1. **One Link Per Student**:
   - Fixed admission number and student details eliminate account-number typos on M-Pesa.
   - Shows live term fee, amount paid so far, and remaining balance.

2. **Supports Flexible Partial Payments**:
   - Parents can pay the full balance, 50%, or a custom installment.
   - Prevents overpaying past the term balance.

3. **M-Pesa STK Push Prompt**:
   - Triggers direct USSD PIN prompts to the parent's phone.
   - Built-in **Demo Mode** simulates live PIN confirmation in 3 seconds for testing without live credentials.
   - Production-ready architecture for Safaricom Daraja API.

4. **Official Digital Fee Receipts**:
   - Real Safaricom transaction reference number (e.g. `SKL4X9Q2TB`).
   - Clean printable/PDF format for parents, complete with verified stamp.
   - Instant WhatsApp sharing.

5. **Bursar Financial Dashboard**:
   - Total expected fees, total collected, total outstanding balance.
   - Clearance rate and breakdown (Cleared, Partial, Owing).
   - Searchable and class-filterable student roster.
   - **"Remind All Owed"**: Send automated fee reminders via SMS/WhatsApp in one click.
   - **CSV Export**: Instant download of the fee ledger for Excel/accounting.

---

## 📁 Project Structure

```
paylink/
├── main.go                     # Server bootstrap & embedded assets
├── go.mod                      # Module definition
├── README.md                   # Project documentation
├── data/
│   └── paylink_db.json         # Persistent JSON fee database
├── internal/
│   ├── config/config.go        # Environment and school settings
│   ├── models/models.go        # Domain models (Student, Payment, School, Reminder)
│   ├── store/store.go          # Thread-safe persistent file store & audit ledger
│   ├── mpesa/mpesa.go          # Safaricom STK Push client & simulation engine
│   ├── sms/sms.go              # SMS reminder & WhatsApp link service
│   └── handlers/
│       ├── routes.go           # HTTP routing table
│       ├── dashboard.go        # School admin handlers & CSV export
│       └── payments.go         # Parent pay portal & receipt views
└── web/
    ├── static/
    │   ├── css/app.css         # Styling (M-Pesa emerald greens & responsive layout)
    │   └── js/app.js           # STK push countdown, copy link, modals, filters
    └── templates/
        ├── landing.html        # Public school-focused landing page
        ├── dashboard.html      # Bursar fee tracking dashboard
        ├── pay.html            # Parent M-Pesa payment link
        └── receipt.html        # Printable official fee receipt
```

---

## 🚀 Quick Start

### 1. Requirements
- Go 1.22 or higher (`go version`)

### 2. Run the App
```bash
go run .
```

The terminal will display:
```
================================================================
 PayLink School Fee Tracker is live at http://localhost:8080
================================================================
 School Name:    Hillcrest Premier Academy
 Current Term:   Term 1, 2026
 School Paybill: 400200
 Mode:           ⚡ Demo / Simulation (Automatic M-Pesa push approval in 3s)
----------------------------------------------------------------
 🏠 Public Landing Page: http://localhost:8080/
 📊 Bursar Dashboard:    http://localhost:8080/admin
 📱 Parent Fee Link:     http://localhost:8080/pay/amani-kimani-adm-2024-012
================================================================
```

### 3. Open in Your Browser
- **Public Landing Page**: [http://localhost:8080](http://localhost:8080)
- **Bursar Admin Dashboard**: [http://localhost:8080/admin](http://localhost:8080/admin)
- **Sample Student Payment Link**: [http://localhost:8080/pay/amani-kimani-adm-2024-012](http://localhost:8080/pay/amani-kimani-adm-2024-012)

---

## ⚙️ Environment Variables (Optional)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port to run the HTTP server |
| `DEMO_MODE` | `true` | Set to `false` for live Daraja / Africa's Talking |
| `SCHOOL_NAME` | `Hillcrest Premier Academy` | Your school's registered name |
| `SCHOOL_PAYBILL` | `400200` | School's Safaricom Paybill or Till number |
| `SCHOOL_CURRENT_TERM` | `Term 1, 2026` | Current academic term |
| `DARAJA_CONSUMER_KEY` | - | Safaricom Daraja Consumer Key |
| `DARAJA_CONSUMER_SECRET` | - | Safaricom Daraja Consumer Secret |
| `DARAJA_PASSKEY` | - | Safaricom Daraja Passkey |
| `AT_API_KEY` | - | Africa's Talking API key for real SMS |
