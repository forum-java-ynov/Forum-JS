function login(event) {
    event.preventDefault();

    fetch("/db/login", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            username: document.getElementById("username").value,
            password: document.getElementById("password").value
        })
    })
    .then(async response => {
        const text = await response.text();

        if (!response.ok) {
            alert(text);
            return;
        }

        alert(text);
        window.location.href = "/";
    });
}
