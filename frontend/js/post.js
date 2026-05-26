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

async function loadComments(postId, commentsContainer) {
    try {
        const response = await fetch(`/db/comments?post_id=${postId}`);

        if (!response.ok) {
            commentsContainer.textContent = "Impossible de charger les commentaires.";
            return;
        }

        const comments = await response.json();
        commentsContainer.innerHTML = "";

        if (!comments || comments.length === 0) {
            return;
        }

        comments.forEach(comment => {
            const commentElement = document.createElement("div");
            commentElement.className = "comment";

            const username = document.createElement("strong");
            username.textContent = comment.username;

            const content = document.createElement("p");
            content.textContent = comment.content;

            commentElement.appendChild(username);
            commentElement.appendChild(content);
            commentsContainer.appendChild(commentElement);
        });
    } catch (error) {
        commentsContainer.textContent = "Impossible de charger les commentaires.";
    }
}

async function loadPosts() {
    const container = document.getElementById("post");

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
            article.dataset.postId = post.id;

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

            const likeBtn = document.createElement("button");

            likeBtn.textContent = `❤️ ${post.likes}`;

            likeBtn.onclick = async function () {
                const response = await fetch(`/db/toggle_like?id=${post.id}`, {
                    method: "POST"
                });

                if (response.ok) {
                    loadPosts();
                } else {
                    alert("Erreur lors du like");
                }
            };
            article.appendChild(likeBtn);

            const commentebutton = document.createElement("button")
            commentebutton.textContent = "Add commente"

            commentebutton.onclick = function(){
                document.getElementById("post-id-commente").value = post.id;
                document.getElementById("create-commente-pop").style.display = "flex";
            }

            const commentsContainer = document.createElement("div");
            commentsContainer.className = "comments";
            article.appendChild(commentsContainer);
            loadComments(post.id, commentsContainer);

            article.appendChild(commentebutton);
            const deleteBtn = document.createElement("button");
            deleteBtn.textContent = "Supprimer";
            deleteBtn.style.backgroundColor = "#ff4c4c";
            deleteBtn.style.color = "white";
            deleteBtn.onclick = () => deletePostAction(post.id);
            
            article.appendChild(deleteBtn);
            container.appendChild(article);
        });
    } catch (error) {
        container.textContent = "Impossible de charger les posts.";
    }
}

async function createCommente(event) {
    event.preventDefault()

    const formData = new FormData();

    formData.append("post_id", document.getElementById("post-id-commente").value);
    formData.append("content", document.getElementById("content-commente").value);

    fetch("/db/create_commente", {
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


async function deletePostAction(postId) {
    if (!postId) {
        alert("Erreur : L'ID du post est introuvable.");
        return;
    }

    if (!confirm("Voulez-vous vraiment supprimer ce post ?")) return;

    const response = await fetch(`/db/delete_post?id=${postId}`, { method: "DELETE" });
    
    if (response.ok) {
        alert("Post supprimé avec succès !");
        loadPosts();
    } else {
        alert("Erreur lors de la suppression du post.");
    }
}

document.addEventListener("DOMContentLoaded", loadPosts);
