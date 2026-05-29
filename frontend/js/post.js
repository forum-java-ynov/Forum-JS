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

            const theme = document.createElement("p");
            theme.textContent = `Thème : ${post.theme}`;

            article.appendChild(title);
            article.appendChild(content);
            article.appendChild(theme);

            if (post.image_path) {
                const image = document.createElement("img");
                image.src = `/uploads/${post.image_path}`;
                image.alt = post.title;
                article.appendChild(image);
            }

            const commentsContainer = document.createElement("div");
            commentsContainer.className = "comments";
            article.appendChild(commentsContainer);
            loadComments(post.id, commentsContainer);

            container.appendChild(article);
        });
    } catch (error) {
        container.textContent = "Impossible de charger les posts.";
    }
}

document.addEventListener("DOMContentLoaded", loadPosts);
*/
