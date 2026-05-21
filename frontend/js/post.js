function createPost(event) {
    event.preventDefault();

    const formData = new FormData();

    

    formData.append("title", document.getElementById("title").value);
    formData.append("content", document.getElementById("content").value);
    
    image = document.getElementById("image_url").files[0]

    if (image) {
        formData.append("image", image)
    }
    
    fetch("/db/create_post", {
        method: "POST",
        body: formData
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

function showPostCreationForm() {
    document.getElementById("posts").style.display = "flex";
}

function closePopup() {
    document.getElementById("posts").style.display = "none";
}