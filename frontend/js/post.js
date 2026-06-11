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

document.addEventListener("DOMContentLoaded", () => {
    const actionForms = document.querySelectorAll('form[action^="/db/toggle_like"], form[action^="/db/toggle_dislike"], form[action^="/db/toggle_comment_like"], form[action^="/db/toggle_comment_dislike"], form[action^="/db/delete_post"]');

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
                }
            } catch (error) {
                console.error("Erreur de réseau :", error);
            }
        });
    });
});