function createPost(event) {
    event.preventDefault();

    const formData = new FormData();

    formData.append("title", document.getElementById("title").value);
    formData.append("content", document.getElementById("content").value);
    
    const image = document.getElementById("image_url").files[0];

    if (image) {
        formData.append("image", image);
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

async function loadPosts() {
    const container = document.getElementById("postContainer");

    try {
        const response = await fetch("/db/posts");
        const posts = await response.json();

        container.innerHTML = "";

        if (!posts || posts.length === 0) {
            const emptyMessage = document.createElement("p");
            emptyMessage.className = "empty-posts";
            emptyMessage.textContent = "Aucun post pour le moment.";
            container.appendChild(emptyMessage);
            return;
        }

        posts.forEach(post => {
            const article = document.createElement("article");
            article.className = "show-post";

            const title = document.createElement("h2");
            title.textContent = post.title;

            const content = document.createElement("p");
            content.textContent = post.content;

            article.appendChild(title);
            article.appendChild(content);

            if (post.image_path) {
                const image = document.createElement("img");
                image.src = `/uploads/${post.image_path}`;
                image.alt = post.title;
                article.appendChild(image);
            }

            container.appendChild(article);
        });
    } catch (error) {
        container.textContent = "Impossible de charger les posts.";
    }
}

document.addEventListener("DOMContentLoaded", loadPosts);

