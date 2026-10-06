// course-nav.js — shared, file://-safe navigation for every Pavona lesson/reference page.
// Keep all lesson entries here so the sidebar stays consistent across the course.
(function () {
  const rootPrefix = /\/(?:lessons|reference)\//.test(window.location.pathname) ? ".." : ".";
  const lessons = [
    ["0001-anatomy-of-a-template.html", "Anatomy of a Template"],
    ["0002-design-pavona-variables.html", "Design Variables That Help"],
    ["0003-install-and-maintain-a-template.html", "Your Template, Available by Name"],
    ["0004-one-input-multiple-representations.html", "One Input, Multiple Representations"],
    ["0005-blueprint-a-template.html", "Blueprint Before You Template"],
    ["0006-diagnose-template-failures.html", "Diagnose by Layer"],
    ["0007-capstone-pavona-workflow.html", "A Repeatable Pavona Workflow"]
  ];

  const currentLessonIndex = lessons.findIndex(function (lesson) {
    return window.location.pathname.endsWith("/" + lesson[0]);
  });

  const sidebar = document.createElement("aside");
  sidebar.className = "course-sidebar";
  sidebar.id = "course-sidebar";
  sidebar.setAttribute("aria-label", "Course navigation");

  const brand = document.createElement("a");
  brand.className = "course-sidebar-brand";
  brand.href = rootPrefix + "/MISSION.html";
  brand.textContent = "Pavona\nA practical course";
  sidebar.appendChild(brand);

  const lessonsHeading = document.createElement("p");
  lessonsHeading.className = "course-sidebar-heading";
  lessonsHeading.textContent = "Lessons";
  sidebar.appendChild(lessonsHeading);

  const list = document.createElement("ol");
  lessons.forEach(function (lesson, index) {
    const item = document.createElement("li");
    const link = document.createElement("a");
    link.className = "course-lesson-link";
    link.href = rootPrefix + "/lessons/" + lesson[0];

    const number = document.createElement("span");
    number.className = "course-lesson-number";
    number.textContent = String(index + 1).padStart(2, "0");

    const title = document.createElement("span");
    title.textContent = lesson[1];
    link.append(number, title);

    if (index === currentLessonIndex) {
      link.setAttribute("aria-current", "page");
    }

    item.appendChild(link);
    list.appendChild(item);
  });
  sidebar.appendChild(list);

  const resourcesHeading = document.createElement("p");
  resourcesHeading.className = "course-sidebar-heading";
  resourcesHeading.textContent = "Quick references";
  sidebar.appendChild(resourcesHeading);

  const references = document.createElement("ul");
  [
    [rootPrefix + "/reference/template-cheatsheet.html", "Template cheat sheet"],
    [rootPrefix + "/reference/template-review-checklist.html", "Review checklist"],
    [rootPrefix + "/MISSION.html", "Learning mission"]
  ].forEach(function (entry) {
    const item = document.createElement("li");
    const link = document.createElement("a");
    link.className = "course-reference-link";
    link.href = entry[0];
    link.textContent = entry[1];
    item.appendChild(link);
    references.appendChild(item);
  });
  sidebar.appendChild(references);

  const footer = document.createElement("p");
  footer.className = "course-sidebar-footer";
  footer.textContent = currentLessonIndex >= 0
    ? "← / p previous · → / n next"
    : "Seven short lessons · Learn by building and checking.";
  sidebar.appendChild(footer);

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.className = "course-nav-toggle";
  toggle.textContent = "☰ Lessons";
  toggle.setAttribute("aria-controls", sidebar.id);
  toggle.setAttribute("aria-expanded", "false");
  toggle.setAttribute("aria-label", "Open course navigation");

  const backdrop = document.createElement("button");
  backdrop.type = "button";
  backdrop.className = "course-nav-backdrop";
  backdrop.setAttribute("aria-label", "Close course navigation");
  backdrop.setAttribute("aria-hidden", "true");
  backdrop.tabIndex = -1;

  document.body.insertBefore(sidebar, document.body.firstChild);
  document.body.insertBefore(backdrop, sidebar.nextSibling);
  document.body.insertBefore(toggle, backdrop.nextSibling);
  document.body.classList.add("course-has-nav");

  const themeControl = document.createElement("div");
  themeControl.className = "course-theme-control";

  const themeToggle = document.createElement("button");
  themeToggle.type = "button";
  themeToggle.className = "course-theme-toggle";
  themeToggle.setAttribute("aria-controls", "course-theme-menu");
  themeToggle.setAttribute("aria-expanded", "false");
  const themeIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  themeIcon.setAttribute("viewBox", "0 0 24 24");
  themeIcon.setAttribute("aria-hidden", "true");
  themeIcon.setAttribute("focusable", "false");
  themeToggle.appendChild(themeIcon);

  const themeMenu = document.createElement("div");
  themeMenu.className = "course-theme-menu";
  themeMenu.id = "course-theme-menu";
  themeMenu.setAttribute("role", "group");
  themeMenu.setAttribute("aria-label", "Color theme");
  themeMenu.hidden = true;

  const themeOptions = [];
  [
    ["light", "Light Mode"],
    ["dark", "Dark Mode"],
    ["system", "System"]
  ].forEach(function (entry) {
    const option = document.createElement("button");
    option.type = "button";
    option.className = "course-theme-option";
    option.textContent = entry[1];
    option.dataset.themeOption = entry[0];
    themeMenu.appendChild(option);
    themeOptions.push(option);
  });

  function updateThemeControl() {
    const mode = window.courseTheme.getMode();
    const labels = { light: "Light Mode", dark: "Dark Mode", system: "System" };
    const shapes = {
      light: [
        ["circle", { cx: "12", cy: "12", r: "3.5" }],
        ["path", { d: "M12 2v2m0 16v2M4.93 4.93l1.42 1.42m11.3 11.3 1.42 1.42M2 12h2m16 0h2M4.93 19.07l1.42-1.42m11.3-11.3 1.42-1.42" }]
      ],
      dark: [["path", { d: "M20.3 15.1A8.5 8.5 0 0 1 8.9 3.7 8.5 8.5 0 1 0 20.3 15.1Z" }]],
      system: [
        ["rect", { x: "3", y: "4", width: "18", height: "13", rx: "2" }],
        ["path", { d: "M8 21h8m-4-4v4" }]
      ]
    };
    const label = "Color theme: " + labels[mode];
    themeToggle.setAttribute("aria-label", label);
    themeToggle.title = label;
    themeIcon.replaceChildren();
    shapes[mode].forEach(function (shape) {
      const element = document.createElementNS("http://www.w3.org/2000/svg", shape[0]);
      Object.keys(shape[1]).forEach(function (attribute) {
        element.setAttribute(attribute, shape[1][attribute]);
      });
      themeIcon.appendChild(element);
    });
    themeOptions.forEach(function (option) {
      option.setAttribute("aria-pressed", String(option.dataset.themeOption === mode));
    });
  }

  function closeThemeMenu(restoreFocus) {
    themeMenu.hidden = true;
    themeToggle.setAttribute("aria-expanded", "false");
    if (restoreFocus) themeToggle.focus();
  }

  themeToggle.addEventListener("click", function () {
    const isOpen = !themeMenu.hidden;
    themeMenu.hidden = isOpen;
    themeToggle.setAttribute("aria-expanded", String(!isOpen));
    if (!isOpen) {
      const selected = themeMenu.querySelector('[aria-pressed="true"]');
      (selected || themeOptions[0]).focus();
    }
  });
  themeMenu.addEventListener("click", function (event) {
    const option = event.target.closest("[data-theme-option]");
    if (!option) return;
    window.courseTheme.setMode(option.dataset.themeOption);
    updateThemeControl();
    closeThemeMenu(true);
  });
  document.body.appendChild(themeControl);
  themeControl.append(themeToggle, themeMenu);
  updateThemeControl();

  const mobile = window.matchMedia("(max-width: 800px)");
  let open = false;

  function setOpen(next, restoreFocus) {
    if (!mobile.matches) return;
    open = next;
    sidebar.classList.toggle("is-open", open);
    document.body.classList.toggle("course-nav-open", open);
    sidebar.inert = !open;
    sidebar.setAttribute("aria-hidden", String(!open));
    backdrop.setAttribute("aria-hidden", String(!open));
    toggle.setAttribute("aria-expanded", String(open));
    toggle.setAttribute("aria-label", open ? "Close course navigation" : "Open course navigation");
    toggle.textContent = open ? "× Close" : "☰ Lessons";

    if (open) {
      const current = sidebar.querySelector('[aria-current="page"]');
      (current || sidebar.querySelector("a")).focus();
    } else if (restoreFocus) {
      toggle.focus();
    }
  }

  function syncViewport() {
    if (mobile.matches) {
      setOpen(false, false);
    } else {
      open = false;
      sidebar.classList.remove("is-open");
      document.body.classList.remove("course-nav-open");
      sidebar.inert = false;
      sidebar.setAttribute("aria-hidden", "false");
      backdrop.setAttribute("aria-hidden", "true");
      toggle.setAttribute("aria-expanded", "false");
      toggle.setAttribute("aria-label", "Open course navigation");
      toggle.textContent = "☰ Lessons";
    }
  }

  toggle.addEventListener("click", function () { setOpen(!open, true); });
  backdrop.addEventListener("click", function () { setOpen(false, true); });
  sidebar.addEventListener("click", function (event) {
    if (event.target.closest("a") && mobile.matches) setOpen(false, false);
  });
  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && !themeMenu.hidden) {
      closeThemeMenu(true);
      return;
    }
    if (event.key === "Escape" && open) {
      setOpen(false, true);
      return;
    }
    if (event.key === "Tab" && open) {
      const focusable = Array.from(sidebar.querySelectorAll("a")).concat(toggle);
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
      return;
    }

    if (currentLessonIndex < 0 || event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey) {
      return;
    }

    const target = event.target;
    if (target.isContentEditable || target.closest("input, textarea, select, button, a, [role='button'], [role='link']")) {
      return;
    }

    const key = event.key.toLowerCase();
    const direction = event.key === "ArrowRight" || key === "n" ? 1
      : event.key === "ArrowLeft" || key === "p" ? -1
      : 0;
    const destination = lessons[currentLessonIndex + direction];
    if (direction === 0 || !destination) return;

    event.preventDefault();
    window.location.href = rootPrefix + "/lessons/" + destination[0];
  });
  document.addEventListener("click", function (event) {
    if (!themeMenu.hidden && !themeControl.contains(event.target)) {
      closeThemeMenu(false);
    }
  });
  if (mobile.addEventListener) mobile.addEventListener("change", syncViewport);
  else mobile.addListener(syncViewport);

  syncViewport();
})();
