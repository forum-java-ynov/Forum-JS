// Toast notification system for auth pages
function showToast(message, type, duration) {
    duration = duration || 4000;
    var container = document.getElementById("toast-container");
    if (!container) return;
    var toast = document.createElement("div");
    toast.className = "toast toast-" + type;
    toast.textContent = message;
    toast.addEventListener("click", function () {
        toast.style.opacity = "0";
        setTimeout(function () { toast.remove(); }, 300);
    });
    container.appendChild(toast);
    setTimeout(function () {
        toast.style.opacity = "0";
        setTimeout(function () { toast.remove(); }, 300);
    }, duration);
}

document.addEventListener("DOMContentLoaded", function () {
    var params = new URLSearchParams(window.location.search);
    if (params.get("error")) {
        showToast(decodeURIComponent(params.get("error")), "error", 5000);
        window.history.replaceState({}, "", window.location.pathname);
    }
    if (params.get("success")) {
        showToast(decodeURIComponent(params.get("success")), "success", 4000);
        window.history.replaceState({}, "", window.location.pathname);
    }
});