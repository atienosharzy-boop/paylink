// Theme Toggle Logic
function initTheme() {
  const savedTheme = localStorage.getItem("paylink_theme") || "light";
  document.documentElement.setAttribute("data-theme", savedTheme);
  updateThemeIcon(savedTheme);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute("data-theme") || "light";
  const next = current === "dark" ? "light" : "dark";
  document.documentElement.setAttribute("data-theme", next);
  localStorage.setItem("paylink_theme", next);
  updateThemeIcon(next);
}

function updateThemeIcon(theme) {
  const icon = document.getElementById("themeIcon");
  const label = document.getElementById("themeLabel");
  if (icon) {
    icon.textContent = theme === "dark" ? "☀️" : "🌙";
  }
  if (label) {
    label.textContent = theme === "dark" ? "Light" : "Dark";
  }
}

// Toast helper
function showToast(message, duration = 3000) {
  let toast = document.getElementById("toastNotification");
  if (!toast) {
    toast = document.createElement("div");
    toast.id = "toastNotification";
    toast.className = "toast";
    document.body.appendChild(toast);
  }
  toast.textContent = message;
  toast.style.display = "block";
  setTimeout(() => {
    toast.style.display = "none";
  }, duration);
}

// Copy to clipboard helper
function copyLink(url) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(url).then(() => {
      showToast("Payment link copied to clipboard!");
    }).catch(() => {
      prompt("Copy this link:", url);
    });
  } else {
    prompt("Copy this link:", url);
  }
}

// Modal management
function openModal(modalId) {
  const el = document.getElementById(modalId);
  if (el) el.classList.add("active");
}

function closeModal(modalId) {
  const el = document.getElementById(modalId);
  if (el) el.classList.remove("active");
}

// Scroll Reveal Observer
function initScrollReveal() {
  const elements = document.querySelectorAll(".reveal");
  if (!elements.length) return;

  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add("active");
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.15 });

  elements.forEach((el) => observer.observe(el));
}

// Counter Animations (e.g. 0 to 60 sec)
function initAnimatedCounters() {
  const counters = document.querySelectorAll("[data-counter-target]");
  if (!counters.length) return;

  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const el = entry.target;
        const target = parseInt(el.getAttribute("data-counter-target"), 10);
        const suffix = el.getAttribute("data-counter-suffix") || "";
        const prefix = el.getAttribute("data-counter-prefix") || "";
        let count = 0;
        const duration = 1500;
        const stepTime = Math.max(20, Math.floor(duration / (target || 1)));

        const timer = setInterval(() => {
          count += Math.ceil(target / 40);
          if (count >= target) {
            count = target;
            clearInterval(timer);
          }
          el.textContent = `${prefix}${count}${suffix}`;
        }, stepTime);

        observer.unobserve(el);
      }
    });
  }, { threshold: 0.3 });

  counters.forEach((c) => observer.observe(c));
}

// Interactive Hero Mini Mockup
function initHeroMockup() {
  const heroChips = document.querySelectorAll(".hero-chip");
  const heroPayBtn = document.getElementById("heroPayBtn");
  const heroNotification = document.getElementById("heroMockupNotification");
  let amount = 10500;

  heroChips.forEach(chip => {
    chip.addEventListener("click", () => {
      heroChips.forEach(c => c.classList.remove("active"));
      chip.classList.add("active");
      amount = parseInt(chip.getAttribute("data-amount"), 10);
      if (heroPayBtn) {
        heroPayBtn.textContent = `Pay KES ${amount.toLocaleString()} with M-Pesa`;
      }
    });
  });

  if (heroPayBtn) {
    heroPayBtn.addEventListener("click", () => {
      heroPayBtn.disabled = true;
      heroPayBtn.innerHTML = `<span>⏳ Processing STK Prompt...</span>`;
      playMpesaTone();

      setTimeout(() => {
        heroPayBtn.disabled = false;
        heroPayBtn.innerHTML = `<span>✅ Paid KES ${amount.toLocaleString()}!</span>`;
        if (heroNotification) {
          heroNotification.style.display = "block";
          heroNotification.innerHTML = `🔔 <strong>M-Pesa Ref SKL4X9Q2TB:</strong> Received KES ${amount.toLocaleString()} for Amani Kimani. Balance updated!`;
        }
        showToast(`M-Pesa payment of KES ${amount.toLocaleString()} confirmed!`);
      }, 2000);
    });
  }
}

// Interactive Savings Calculator
function initCalculator() {
  const slider = document.getElementById("calcStudents");
  const studentsVal = document.getElementById("calcStudentsVal");
  const hoursSaved = document.getElementById("calcHoursSaved");
  const feesProtected = document.getElementById("calcFeesProtected");

  if (!slider) return;

  function update() {
    const students = parseInt(slider.value, 10);
    if (studentsVal) studentsVal.textContent = `${students} students`;
    
    // Calculate realistic savings
    // Approx 1 hour per 15 students spent verifying screenshots/term
    const hours = Math.round((students * 1.8) / 10);
    // Approx KES 1,200 per student in avoided dispute/lost track
    const fees = students * 1400;

    if (hoursSaved) hoursSaved.textContent = `${hours} hrs / term`;
    if (feesProtected) feesProtected.textContent = `KES ${fees.toLocaleString()}`;
  }

  slider.addEventListener("input", update);
  update();
}

// FAQ Accordion
function initFAQ() {
  const items = document.querySelectorAll(".faq-item");
  items.forEach(item => {
    const q = item.querySelector(".faq-question");
    if (q) {
      q.addEventListener("click", () => {
        const isActive = item.classList.contains("active");
        items.forEach(i => i.classList.remove("active"));
        if (!isActive) {
          item.classList.add("active");
        }
      });
    }
  });
}

// Send single SMS reminder from dashboard
async function sendSingleReminder(studentId) {
  try {
    showToast("Sending reminder SMS...");
    const res = await fetch(`/api/students/${studentId}/remind`, {
      method: "POST",
      headers: { "Accept": "application/json" }
    });
    const data = await res.json();
    if (!res.ok) {
      alert(data.error || "Failed to send reminder");
      return;
    }
    showToast("SMS reminder sent to parent!");
  } catch (err) {
    console.error(err);
    alert("Network error sending reminder");
  }
}

// Send reminders to all students with balances
async function remindAllOwed() {
  if (!confirm("Are you sure you want to send SMS fee reminders to ALL parents with outstanding balances?")) {
    return;
  }
  try {
    showToast("Dispatching reminders...");
    const res = await fetch("/api/remind-all", {
      method: "POST",
      headers: { "Accept": "application/json" }
    });
    const data = await res.json();
    if (res.ok) {
      alert(`Success: ${data.message}`);
    } else {
      alert("Error: " + (data.error || "Could not dispatch reminders"));
    }
  } catch (err) {
    console.error(err);
    alert("Network error processing bulk reminders");
  }
}

// Delete student confirmation
async function deleteStudent(studentId, name) {
  if (!confirm(`Are you sure you want to remove ${name} from the fee tracker?`)) {
    return;
  }
  try {
    const res = await fetch(`/api/students/${studentId}/delete`, {
      method: "POST",
      headers: { "Accept": "application/json" }
    });
    if (res.ok) {
      window.location.reload();
    } else {
      alert("Failed to delete student record");
    }
  } catch (err) {
    console.error(err);
  }
}

// Simple Web Audio API beep for M-Pesa prompt simulation
function playMpesaTone() {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = "sine";
    osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
    gain.gain.setValueAtTime(0.2, ctx.currentTime);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + 0.25);
  } catch (e) {
    // AudioContext blocked or not supported
  }
}

// Parent Payment Link Logic
function initPaymentFlow(studentId, balance) {
  let selectedAmount = balance;
  const chips = document.querySelectorAll(".chip");
  const customAmountInput = document.getElementById("customAmount");
  const payBtn = document.getElementById("payMpesaBtn");
  const phoneInput = document.getElementById("parentPhone");

  function updatePayBtn(amount) {
    selectedAmount = amount;
    if (payBtn) {
      payBtn.textContent = `Pay KES ${amount.toLocaleString()} with M-Pesa`;
      payBtn.disabled = amount <= 0 || amount > balance;
    }
  }

  chips.forEach(chip => {
    chip.addEventListener("click", () => {
      chips.forEach(c => c.classList.remove("active"));
      chip.classList.add("active");
      const val = parseInt(chip.getAttribute("data-amount"), 10);
      if (customAmountInput) customAmountInput.value = "";
      updatePayBtn(val);
    });
  });

  if (customAmountInput) {
    customAmountInput.addEventListener("input", (e) => {
      chips.forEach(c => c.classList.remove("active"));
      const val = parseInt(e.target.value, 10) || 0;
      updatePayBtn(val);
    });
  }

  if (payBtn) {
    payBtn.addEventListener("click", async () => {
      const phone = phoneInput ? phoneInput.value.trim() : "";
      if (!phone) {
        alert("Please enter a valid M-Pesa phone number");
        return;
      }

      if (selectedAmount <= 0 || selectedAmount > balance) {
        alert(`Payment must be between KES 1 and your balance of KES ${balance.toLocaleString()}`);
        return;
      }

      openModal("stkModal");
      playMpesaTone();

      try {
        const initRes = await fetch("/api/pay/stk-push", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            student_id: studentId,
            amount: selectedAmount,
            phone_number: phone
          })
        });

        const initData = await initRes.json();
        if (!initRes.ok) {
          closeModal("stkModal");
          alert(initData.error || "STK Push failed");
          return;
        }

        let counter = 3;
        const countdownEl = document.getElementById("stkCountdown");
        const interval = setInterval(() => {
          counter--;
          if (countdownEl) countdownEl.textContent = counter;
          if (counter <= 0) {
            clearInterval(interval);
            finalizePayment(studentId, selectedAmount, phone, initData.receipt_number);
          }
        }, 1000);

      } catch (err) {
        closeModal("stkModal");
        alert("Error connecting to payment gateway: " + err.message);
      }
    });
  }

  async function finalizePayment(studentId, amount, phone, receiptNumber) {
    try {
      const res = await fetch("/api/pay/confirm", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          student_id: studentId,
          amount: amount,
          phone_number: phone,
          receipt_number: receiptNumber,
          notes: "Official Fee Payment via PayLink"
        })
      });

      const data = await res.json();
      if (res.ok && data.receipt_url) {
        window.location.href = data.receipt_url;
      } else {
        alert("Payment confirmation error: " + (data.error || "Unknown"));
        closeModal("stkModal");
      }
    } catch (err) {
      alert("Error confirming payment: " + err.message);
      closeModal("stkModal");
    }
  }
}

// Global initialization
document.addEventListener("DOMContentLoaded", () => {
  initTheme();
  initScrollReveal();
  initAnimatedCounters();
  initHeroMockup();
  initCalculator();
  initFAQ();
});
