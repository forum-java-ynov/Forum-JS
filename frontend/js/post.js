function showPostCreationForm() {
    const popup = document.getElementById("posts");
    if (popup) popup.style.display = "flex";
}

function closePopup() {
    const popup = document.getElementById("posts");
    if (popup) popup.style.display = "none";
}

function showCommenteCreationForm(postId) {
    const popup = document.getElementById("create-commente-pop");
    const postInput = document.getElementById("post-id-commente");

    if (postInput) postInput.value = postId;
    if (popup) popup.style.display = "flex";
}

function closeCommentePopup() {
    const popup = document.getElementById("create-commente-pop");
    if (popup) popup.style.display = "none";
}

/*
async function loadComments(postId, commentsContainer) {
    try {
        const response = await fetch(`/db/comments?post_id=${postId}`);
        if (!response.ok) {
            commentsContainer.textContent =
                "Impossible de charger les commentaires.";
            return;
        }

        const comments = await response.json();
        commentsContainer.innerHTML = "";

        if (!comments || comments.length === 0) return;

        comments.forEach(comment => {
            const item = document.createElement("div");
            item.className = "comment-item";

            const left = document.createElement("div");

            const username = document.createElement("strong");
            username.textContent = comment.username;

            const content = document.createElement("p");
            content.textContent = comment.content;

            commentElement.appendChild(username);
            commentElement.appendChild(content);
            commentsContainer.appendChild(commentElement);
        });

    } catch (err) {
        commentsContainer.textContent =
            "Impossible de charger les commentaires.";
    }
}

/* posts */

async function loadPosts() {
    const container = document.getElementById("postContainer");

    if (!container) return;

    try {
        const response = await fetch("/db/posts");
        if (!response.ok) {
            container.textContent = "Impossible de charger les posts.";
            return;
        }

        const posts = await response.json();
        container.innerHTML = "";

        if (!posts || posts.length === 0) {
            const empty = document.createElement("p");
            empty.className = "empty-state";
            empty.textContent = "Aucun post pour le moment.";

            container.appendChild(empty);
            return;
        }

        posts.forEach(post => {
            const article = document.createElement("article");
            article.className = "show-post";

            if (post.image_path) {
                const img = document.createElement("img");
                img.src = `/uploads/${post.image_path}`;
                img.alt = post.title;
                img.className = "post-card-img";

                article.appendChild(img);
            }

            const body = document.createElement("div");
            body.className = "post-card-body";

            const title = document.createElement("h3");
            title.textContent = post.title;

            const content = document.createElement("p");
            content.textContent = post.content;

            const theme = document.createElement("p");
            theme.textContent = `Thème : ${post.theme}`;

            article.appendChild(title);
            article.appendChild(content);
            article.appendChild(theme);

            article.appendChild(body);

            const commentsContainer = document.createElement("div");
            commentsContainer.className = "comments";
            article.appendChild(commentsContainer);
            loadComments(post.id, commentsContainer);

            container.appendChild(article);
        });

    } catch (err) {
        container.textContent = "Impossible de charger les posts.";
    }
}

document.addEventListener("DOMContentLoaded", loadPosts);
*/
