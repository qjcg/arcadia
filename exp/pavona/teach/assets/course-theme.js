// course-theme.js — apply the saved course theme before stylesheets paint.
(function () {
  const storageKey = "pavona-course-theme";
  const validModes = ["light", "dark", "system"];
  let mode = "system";

  try {
    const saved = window.localStorage.getItem(storageKey);
    if (validModes.includes(saved)) mode = saved;
  } catch (_) {
    // Storage may be unavailable for file:// pages; keep the system default.
  }

  function apply(nextMode) {
    mode = validModes.includes(nextMode) ? nextMode : "system";
    document.documentElement.dataset.theme = mode;
    try {
      window.localStorage.setItem(storageKey, mode);
    } catch (_) {
      // Theme still applies for this page even if it cannot be persisted.
    }
  }

  document.documentElement.dataset.theme = mode;
  window.courseTheme = {
    getMode: function () { return mode; },
    setMode: apply
  };
})();
