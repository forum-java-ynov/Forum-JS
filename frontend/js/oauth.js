document.addEventListener("DOMContentLoaded", () => {
    fetch('/api/me')
        .then(r => r.ok ? r.json() : null)
        .then(data => {
            if (!data) return;

            document.getElementById('nav-login').style.display = 'none'
            document.getElementById('nav-register').style.display = 'none'
            document.getElementById('nav-user').style.display = 'flex'
            document.getElementById('nav-logout').style.display = 'block'

            const nameEl = document.getElementById("nav-name");
            const picEl = document.getElementById("nav-pfp");

            if (nameEl) {
                nameEl.textContent = data.name || data.email;
            }

            if (picEl && data.picture) {
                picEl.src = data.picture;
            }

            const googleBtn = document.querySelector('a[href="/auth/google/login"]')
            if (googleBtn) googleBtn.style.display = 'none'
        })
})