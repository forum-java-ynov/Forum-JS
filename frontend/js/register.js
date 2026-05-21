function register(event) {
    event.preventDefault();

    fetch("/db/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            full_name: document.getElementById("full_name").value,
            username: document.getElementById("username").value,
            email: document.getElementById("email").value,
            password: document.getElementById("password").value,
            confirm_password: document.getElementById("confirm_password").value
        })
    })
    .then(async response => {
    const text = await response.text();

    if (!response.ok) {
        alert(text);
        return;
    }

    alert(text);
    window.location.href = "/login";
})
}