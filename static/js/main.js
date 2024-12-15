Prism.plugins.NormalizeWhitespace.setDefaults({
  'remove-trailing': true,
  'remove-indent': true,
  'left-trim': true,
  'right-trim': false,
  'remove-initial-line-feed': true,
  'tabs-to-spaces': 4,
  'spaces-to-tabs': 4
});

// Make the header sticky on scroll
window.addEventListener("scroll", () => {
  const header = document.querySelector("header")
  if (header === null) return
  // If the user scrolls down, add "sticky" class
  if (window.scrollY > 0) {
    header.classList.add("sticky")
  } else { // The user is at the top of the page, remove class
    header.classList.remove("sticky")
  }
}, { passive: true })

htmx.onLoad((content) => {
  // Always call hideHamMenu
  hideHamMenu();

  // Check if the current page URL matches "/showcase/article" and call Prism.highlightAll
  if (window.location.pathname === "/showcase/article") {
    Prism.highlightAll();
  }

  // Check if the current page URL contains "/download" and call handleDropdowns
  if (window.location.pathname.includes("/download")) {
    handleDropdowns(content);
  }

  // Check if the current page URL contains "/road-maps" and load embeds
  if (window.location.pathname.includes("/road-maps")) {
    const embeds = ["miro", "timeGraphics"];

    embeds.forEach(embed => {
      loadEmbed(embed)
    });
  }
});

function showHamMenu() {
  const menu = document.querySelector("#hamburger-nav")
  menu.style.display = "flex"
}

function hideHamMenu() {
  const menu = document.querySelector("#hamburger-nav")
  menu.style.display = "none"
}

function allowEmbed(embedName) {
  setConsentCookie(embedName)
  loadEmbed(embedName);
}

function setConsentCookie(embedName) {
  const expiryDate = new Date();
  expiryDate.setFullYear(expiryDate.getFullYear() + 1); // Cookie expires in 1 year
  document.cookie = `${embedName}Consent=true; expires=${expiryDate.toUTCString()}; path=/`;
}

function checkConsent(embedName) {
  const consent = document.cookie.split('; ').find(row => row.startsWith(`${embedName}Consent=`));
  return consent ? consent.split('=')[1] === 'true' : false;
}

function loadEmbed(embedName) {
  if (checkConsent(embedName)) {
    document.getElementById(`${embedName}-placeholder`).style.display = 'none';
    document.getElementById(`${embedName}-embed`).style.display = 'block';
  }
}

function deleteConsentCookies() {
  const consentCookies = ["miro", "timeGraphics"];

  consentCookies.forEach(cookie => {
    document.cookie = `${cookie}Consent=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/`;
  });

  alert("Consent cookies have been deleted!");
}

function handleDropdowns(content) {
  // Helper function to fetch options
  const fetchOptions = async () => {
    try {
      const response = await fetch("/download-options", {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      });
      if (!response.ok) {
        throw new Error(`Failed to fetch options: HTTP ${response.status}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching options:', error);
      return null;
    }
  };


  // Helper function to set links and text after selecting an option
  var commands
  const setLinks = () => {
    const dropdowns = content.querySelectorAll(".dropdown");
    const options = {};

    // Collect selected options from dropdowns
    dropdowns.forEach(dropdown => {
      const selectedItem = dropdown.querySelector(".selected");
      options[dropdown.id] = selectedItem.textContent;
    });

    // Update the download button or command dynamically
    const downloadButton = content.querySelector("#download-btn");
    const downloadCmd = content.querySelector("#download-cmd");

    if (downloadButton) {
      const version = options.version;
      downloadButton.href = `/${version}`;
      const buttonLabel = downloadButton.querySelector("span");
      if (buttonLabel) {
        buttonLabel.textContent = version;
      }
    }

    if (downloadCmd) {
      // Choose the correct command based on the selected package manager and version from JSON
      const version = options.version;
      const pkgmngr = options.pkgmngr;
      const commandTemplate = commands[pkgmngr];
      const cmd = commandTemplate ? commandTemplate.replace("{version}", version) : "";

      downloadCmd.querySelector("code").textContent = cmd;
      Prism.highlightAll()
    }
  };

  // Helper function to create dropdown menu items
  const createMenuItems = (menu, options, button) => {
    options.forEach(item => {
      // Set button text to item text to find the minimum width to fit the content
      button.querySelector(".text").textContent = item
      menu.style.minWidth = `${button.offsetWidth}px`
      button.style.minWidth = `${menu.offsetWidth}px`

      const itemElement = document.createElement('li');
      itemElement.textContent = item;
      menu.appendChild(itemElement);

      // Add item click event
      itemElement.addEventListener("click", () => {
        selectItem(itemElement, button, menu);
      });
    });

    // Set initial selection
    if (options.length > 0) {
      const firstItem = menu.querySelector("li");
      firstItem.classList.add("selected");
      button.querySelector(".text").textContent = firstItem.textContent;
    }
  };

  // Helper function to handle item selection
  const selectItem = (item, button, menu) => {
    // Deselect all items
    menu.querySelectorAll("li").forEach(i => i.classList.remove("selected"));

    // Select current item
    item.classList.add("selected");
    button.querySelector(".text").textContent = item.textContent;

    // Close dropdown
    menu.classList.remove("active");

    // Update links
    setLinks()
  };

  // Fetch and populate dropdowns
  fetchOptions().then(data => {
    if (!data) return; // Exit if fetch failed

    commands = data.commands

    const dropdowns = content.querySelectorAll(".dropdown");
    dropdowns.forEach(dropdown => {
      const button = dropdown.querySelector(".dropdown-button");
      const menu = dropdown.querySelector(".dropdown-menu");

      // Add dropdown toggle event
      button.addEventListener("click", () => menu.classList.toggle("active"));

      // Populate menu items
      const optionList = data.options[dropdown.id] || [];
      createMenuItems(menu, optionList, button);
    });

    // Close dropdowns when clicking outside
    document.addEventListener("click", (event) => {
      dropdowns.forEach(dropdown => {
        if (!dropdown.contains(event.target)) {
          const menu = dropdown.querySelector(".dropdown-menu");
          menu.classList.remove("active");
        }
      });
    });

    // Inital setup
    setLinks()
  });
}
