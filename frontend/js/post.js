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

function showEditForm(commentId) {
    const popup = document.getElementById("edit-comment-pop");
    const commentInput = document.getElementById("edit-comment-id");
    const contentArea = document.getElementById("edit-content-commente");
    const currentContent = document.getElementById("comment-content-" + commentId);

    if (commentInput) commentInput.value = commentId;
    if (contentArea && currentContent) contentArea.value = currentContent.innerText;

    if (popup) popup.style.display = "flex";
}

function closeCommentePopup() {
    const popup = document.getElementById("create-commente-pop");
    if (popup) popup.style.display = "none";
}

function closeEditPopup() {
    const popup = document.getElementById("edit-comment-pop");
    if (popup) popup.style.display = "none";
}

function showEditPostForm(postId) {
    const popup = document.getElementById("edit-post-pop");
    const idInput = document.getElementById("edit-post-id");
    const titleInput = document.getElementById("edit-post-title");
    const contentArea = document.getElementById("edit-post-content");

    const currentTitle = document.getElementById("post-title-" + postId);
    const currentContent = document.getElementById("post-content-" + postId);

    if (idInput) idInput.value = postId;
    if (titleInput && currentTitle) titleInput.value = currentTitle.innerText;
    if (contentArea && currentContent) contentArea.value = currentContent.innerText;

    if (popup) popup.style.display = "flex";
}

function closeEditPostPopup() {
    const popup = document.getElementById("edit-post-pop");
    if (popup) popup.style.display = "none";
}

function filterPosts() {
    const select = document.getElementById("filter-theme-select");
    if (!select) return;
    const theme = select.value;
    
    if (theme === "") {
        window.location.href = "/";
    } else {
        window.location.href = "/?theme=" + encodeURIComponent(theme);
    }
}

function resetFilter() {
    window.location.href = "/";
}

document.addEventListener("DOMContentLoaded", () => {
    const actionForms = document.querySelectorAll('form[action^="/db/toggle_like"], form[action^="/db/toggle_dislike"], form[action^="/db/toggle_comment_like"], form[action^="/db/toggle_comment_dislike"], form[action^="/db/delete_post"], form[action*="edit"]');

    actionForms.forEach(form => {
        form.addEventListener('submit', async (event) => {
            event.preventDefault();
            
            const url = form.getAttribute('action');
            const formData = new FormData(form);
            
            try {
                const response = await fetch(url, {
                    method: 'POST',
                    body: formData,
                    headers: {
                        'Accept': 'application/json'
                    }
                });
                
                if (response.ok) {
                    location.reload();
                } else {
                    console.error("Erreur HTTP:", response.status);
                    alert("Erreur lors de la modification. Le serveur a renvoyé le code : " + response.status + "\n\nAvez-vous bien ajouté la route /api/edit-post dans votre code Go ?");
                }
            } catch (error) {
                console.error("Erreur de réseau :", error);
                alert("Erreur de réseau. Le serveur Go est-il bien lancé ?");
            }
        });
    });
    const urlParams = new URLSearchParams(window.location.search);
    const activeTheme = urlParams.get('theme');
    const filterBtn = document.getElementById("filter-btn");
    if (activeTheme && filterBtn) {
        filterBtn.textContent = "Reset Filter";
        filterBtn.onclick = resetFilter;
    }
});