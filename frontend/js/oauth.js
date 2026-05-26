fetch('/api/me')
    .then(r => r.ok ? r.json() : null)
    .then(data => {
        if (data) {
            document.getElementById('nav-login').style.display = 'none'
            document.getElementById('nav-register').style.display = 'none'
            document.getElementById('nav-user').style.display = 'flex'
            document.getElementById('nav-logout').style.display = 'block'
            document.getElementById('nav-email').textContent = data.email
            if (data.picture) {
                document.getElementById('nav-pfp').src = data.picture
            }
            const googleBtn = document.querySelector('a[href="/auth/google/login"]')
            if (googleBtn) googleBtn.style.display = 'none'
        }
    })
