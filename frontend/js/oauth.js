document.addEventListener("DOMContentLoaded", () => {
    fetch('/api/me')
        .then(async r => {
            if (!r.ok) return null;
            return await r.json();
        })
        .then(data => {
            if (!data) return;

            // hide login/register
            const login = document.getElementById('nav-login');
            const register = document.getElementById('nav-register');

            if (login) login.style.display = 'none';
            if (register) register.style.display = 'none';

            // show user
            const user = document.getElementById('nav-user');
            const logout = document.getElementById('nav-logout');

            if (user) user.style.display = 'flex';
            if (logout) logout.style.display = 'block';

            // name
            const nameEl = document.getElementById("nav-name");
            if (nameEl) {
                nameEl.textContent = data.name || data.email;
            }

            // picture
            const picEl = document.getElementById("nav-pfp");
            if (picEl) {
                picEl.src = data.picture || "";
            }

            // hide google button
            const googleBtn = document.querySelector('a[href="/auth/google/login"]');
            if (googleBtn) googleBtn.style.display = 'none';
        });
});