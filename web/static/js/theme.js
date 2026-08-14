// Theme toggle. The initial theme is applied by an inline <head> script
// (stored choice, else prefers-color-scheme) to avoid a flash. This only
// labels the button and persists on explicit user toggle — so a visitor who
// never toggles keeps following their OS preference.
(function () {
  function current() {
    return document.documentElement.getAttribute("data-theme") || "dark";
  }
  function label(theme) {
    var btn = document.getElementById("theme-toggle");
    if (btn) btn.textContent = theme === "dark" ? "◐ light" : "◑ dark";
  }
  function setTheme(theme) {
    document.documentElement.setAttribute("data-theme", theme);
    try { localStorage.setItem("theme", theme); } catch (e) {}
    label(theme);
  }
  document.addEventListener("DOMContentLoaded", function () {
    label(current());
    var btn = document.getElementById("theme-toggle");
    if (btn) {
      btn.addEventListener("click", function () {
        setTheme(current() === "dark" ? "light" : "dark");
      });
    }
    requestAnimationFrame(function () {
      document.body.classList.remove("notransition");
    });
  });
})();
