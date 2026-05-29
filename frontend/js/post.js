function showPostCreationForm() {
    document.getElementById("posts").classList.add("visible");
}

function closePopup() {
    document.getElementById("posts").classList.remove("visible");
}

function closeCommentPopup() {
    document.getElementById("create-commente-pop").classList.remove("visible");
}

/* create post */

async function createPost(event) {
    event.preventDefault();

    const formData = new FormData();

    formData.append(
        "title",
        document.getElementById("title").value.trim()
    );

    formData.append(
        "content",
        document.getElementById("content").value.trim()
    );

    const image = document.getElementById("image_url").files[0];

    if (image) {
        formData.append("image", image);
    }

    try {
        const response = await fetch("/db/create_post", {
            method: "POST",
            body: formData
        });

        const text = await response.text();

        if (!response.ok) {
            alert(text);
            return;
        }

        closePopup();
        document.getElementById("postForm").reset();

        await loadPosts();

    } catch (err) {
        alert("Erreur lors de la création du post.");
    }
}

/* comments */

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

            left.appendChild(username);
            left.appendChild(content);

            const likeBtn = document.createElement("button");
            likeBtn.className = "comment-like-btn";
            likeBtn.textContent = `❤️ ${comment.likes}`;

            likeBtn.onclick = async () => {
                try {
                    const res = await fetch(
                        `/db/toggle_comment_like?id=${comment.id}`,
                        { method: "POST" }
                    );

                    if (res.ok) {
                        loadComments(postId, commentsContainer);
                    }
                } catch (e) {
                    alert("Erreur lors du like.");
                }
            };

            item.appendChild(left);
            item.appendChild(likeBtn);

            commentsContainer.appendChild(item);
        });

    } catch (err) {
        commentsContainer.textContent =
            "Impossible de charger les commentaires.";
    }
}

/* posts */

async function loadPosts() {
    const container = document.getElementById("post");

    try {
        const response = await fetch("/db/posts");

        if (!response.ok) throw new Error();

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
            article.className = "post-card";
            article.dataset.postId = post.id;

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

            body.appendChild(title);
            body.appendChild(content);

            article.appendChild(body);

            const comments = document.createElement("div");
            comments.className = "post-comments-wrapper";

            article.appendChild(comments);
            loadComments(post.id, comments);

            const footer = document.createElement("div");
            footer.className = "post-card-footer";

            const likeBtn = document.createElement("button");
            likeBtn.className = "post-action-btn";
            likeBtn.textContent = `❤️ ${post.likes}`;

            likeBtn.onclick = async () => {
                try {
                    const res = await fetch(
                        `/db/toggle_like?id=${post.id}`,
                        { method: "POST" }
                    );

                    if (res.ok) loadPosts();
                } catch (e) {
                    alert("Erreur lors du like.");
                }
            };

            const commentBtn = document.createElement("button");
            commentBtn.className = "post-comment-btn";
            commentBtn.textContent = "💬 Commenter";

            commentBtn.onclick = () => {
                document.getElementById("post-id-commente").value = post.id;
                document.getElementById("create-commente-pop").classList.add("visible");
            };

            const deleteBtn = document.createElement("button");
            deleteBtn.className = "post-delete-btn";
            deleteBtn.textContent = "Supprimer";

            deleteBtn.onclick = () => deletePostAction(post.id);

            footer.appendChild(likeBtn);
            footer.appendChild(commentBtn);
            footer.appendChild(deleteBtn);

            article.appendChild(footer);
            container.appendChild(article);
        });

    } catch (err) {
        container.textContent = "Impossible de charger les posts.";
    }
}

/* create comment */

async function createCommente(event) {
    event.preventDefault();

    const formData = new FormData();

    formData.append(
        "post_id",
        document.getElementById("post-id-commente").value
    );

    formData.append(
        "content",
        document.getElementById("content-commente").value.trim()
    );

    try {
        const response = await fetch("/db/create_commente", {
            method: "POST",
            body: formData
        });

        const text = await response.text();

        if (!response.ok) {
            alert(text);
            return;
        }

        closeCommentPopup();
        document.getElementById("commenteForm").reset();

        await loadPosts();

    } catch (err) {
        alert("Erreur lors du commentaire.");
    }
}

/* delete a post */

async function deletePostAction(postId) {
    if (!postId) {
        alert("Erreur : ID introuvable.");
        return;
    }

    if (!confirm("Voulez-vous vraiment supprimer ce post ?")) return;

    try {
        const response = await fetch(
            `/db/delete_post?id=${postId}`,
            { method: "DELETE" }
        );

        if (response.ok) {
            loadPosts();
        } else {
            alert("Erreur lors de la suppression.");
        }

    } catch (err) {
        alert("Erreur serveur.");
    }
}

/* events */

window.addEventListener("click", (e) => {
    const postModal = document.getElementById("posts");
    const commentModal = document.getElementById("create-commente-pop");

    if (e.target === postModal) closePopup();
    if (e.target === commentModal) closeCommentPopup();
});

document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
        closePopup();
        closeCommentPopup();
    }
});




document.addEventListener("DOMContentLoaded", loadPosts);