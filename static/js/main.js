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

    const handleLink = (elementId) => {
      const element = content.querySelector(`#${elementId}`);
      if (element) {
        element.addEventListener("click", async (event) => {
          event.preventDefault();

          const href = element.getAttribute("href");
          try {
            const response = await fetch(href, { method: "HEAD"});
            if (response.status === 404) {
              alert("The file does not exist (404).");
            } else {
              window.location.href = href;
            }
          } catch (error) {
            alert("An error occurred while checking the file.");
          }
        })
      }
    }
    handleLink("download-btn");
    handleLink("templates");
    handleLink("templates-mono");

    // handle changelog button
    const changelogButton = content.querySelector("#changelog-btn");
    if (changelogButton) {
      changelogButton.addEventListener("click", async (event) => {
        event.preventDefault();

        const href = changelogButton.getAttribute("href");
        try {
          var link = `https://api.github.com/repos/blazium-engine/blazium/releases/tags/${href}`
          const response = await fetch(link, { method: "HEAD"});
          if (response.status === 404) {
            alert("No changelog is present for this version.");
          } else {
            link = `https://github.com/blazium-engine/blazium/releases/tag/${href}`
            window.location.href = link;
          }
        } catch (error) {
          alert("An error occurred while checking the link.");
        }
      })
    }
  }

  // Check if the current page URL contains "/road-maps" and load embeds
  if (window.location.pathname === "/road-maps") {
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
  var versions;
  var options;
  var commands;

  // Helper function to fetch options
  const fetchOptions = async () => {
    try {
      const response = await fetch("/api/download-options", {
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

  const getSystemInfo = () => {
    const userAgent = window.navigator.userAgent.toLowerCase();

    let os = undefined;
    if (/windows/.test(userAgent)) {
      os = "Windows";
    } else if (/mac/.test(userAgent)) {
      os = "MacOS";
    } else if (/linux/.test(userAgent)) {
      os = "Linux";
    } else if (/android/.test(userAgent)) {
      os = "Android";
    // } else if (/iPhone|iPad|iPod/.test(userAgent)) {
    //   os = "ios";
    }

    let arch = undefined
    if (/arm64|aarch64/.test(userAgent)) {
      arch = "ARM64";
    } else if (/arm|aarch32/.test(userAgent)) {
      arch = "ARM32";
    } else if (/x86_64|win64|wow64/.test(userAgent)) {
      arch = "x86_64";
    } else if (/x86|win32/.test(userAgent)) {
      arch = "x86_32";
    }

    return {"os": os, "arch": arch};
  }

  // Helper function to set links and text after selecting an option
  const setLinks = (firstLoad=false) => {
    if (firstLoad) {
      const {os, arch} = getSystemInfo();

      if (os) {
        const osDropdown = content.querySelector("#os");
        const osDropdownButton = osDropdown.querySelector(".dropdown-button");
        const osDropdownMenu = osDropdown.querySelector(".dropdown-menu");
        const itemToSelect = osDropdownMenu.querySelector("#"+os);

        selectItem(itemToSelect, osDropdownButton, osDropdownMenu, false);
      }
      if (arch) {
        const archDropdown = content.querySelector("#arch");
        const archDropdownButton = archDropdown.querySelector(".dropdown-button");
        const archDropdownMenu = archDropdown.querySelector(".dropdown-menu");
        const itemToSelect = archDropdownMenu.querySelector("#"+arch);

        selectItem(itemToSelect, archDropdownButton, archDropdownMenu, false);
      }
    }

    // Collect selected options from dropdowns
    const dropdowns = content.querySelectorAll(".dropdown");
    const selectedOptions = {};
    dropdowns.forEach(dropdown => {
      const selectedItem = dropdown.querySelector(".selected");
      selectedOptions[dropdown.id] = selectedItem.textContent;
    });

    // Hide arch dropdown when macOS
    const macosToHide = content.querySelector("#no-macos");
    if (macosToHide) {
      if (selectedOptions.os === "MacOS") {
        macosToHide.style.display = "none";
      } else if (macosToHide.style.display === "none") {
        macosToHide.style.display = "inline-block";
      }
    }

    // Hide arch and mono dropdown when andoid
    const androidToHide = content.querySelector("#no-android");
    if (androidToHide) {
      if (selectedOptions.os === "Android" || selectedOptions.os === "HorizonOS") {
        androidToHide.style.display = "none";
      } else if (androidToHide.style.display === "none") {
        androidToHide.style.display = "inline-block";
      }
    }

    // Update the links dynamically
    const changelogButton = content.querySelector("#changelog-btn");
    const templatesContainer = content.querySelector("#export-templates");
    const downloadButton = content.querySelector("#download-btn");
    const downloadCmd = content.querySelector("#download-cmd");

    if (changelogButton) {
      const version = selectedOptions.version;
      const buildType = selectedOptions.buildType;
      changelogButton.href = `v${version}-${buildType}`
    }

    if (templatesContainer) {
      const templates = templatesContainer.querySelector("#templates");
      const templatesMono = templatesContainer.querySelector("#templates-mono");

      const version = selectedOptions.version;
      const buildType = selectedOptions.buildType;

      const textContent = `${version} ${buildType}`

      templates.href = `https://cdn.blazium.app/${buildType}/${version}/Blazium_v${version}_export_templates.tpz`
      const templatesLabel = templates.querySelector("span");
      if (templatesLabel) {
        templatesLabel.textContent = textContent;
      }
      templatesMono.href = `https://cdn.blazium.app/${buildType}/${version}/Blazium_v${version}_mono_export_templates.tpz`
      const templatesMonoLabel = templatesMono.querySelector("span");
      if (templatesMonoLabel) {
        templatesMonoLabel.textContent = `${textContent} - .NET/C#`;
      }
    }

    if (downloadButton) {
      const version = selectedOptions.version;
      const buildType = selectedOptions.buildType;
      const os = selectedOptions.os.toLowerCase().replace(/\s+/g, '');
      let arch = "." + selectedOptions.arch.toLowerCase();
      if (os === "windows" && arch.includes("x86")) {
        arch = arch === ".x86_64" ? ".64bit" : ".32bit"
      } else if (os !== "linux") {
        arch = ""
      }

      const isMono =  selectedOptions.csharp === "with" ? ".mono" : ""

      const link = `https://cdn.blazium.app/${buildType}/${version}/BlaziumEditor_v${version}_${os}${isMono}${arch}.zip`
      downloadButton.href = link;

      const buttonLabel = downloadButton.querySelector("span");
      if (buttonLabel) {
        buttonLabel.textContent = version;
      }
    }

    if (downloadCmd) {
      // Choose the correct command based on the selected package manager and version from JSON
      const version = selectedOptions.version;
      const pkgmngr = selectedOptions.pkgmngr;
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
      itemElement.id = item;
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

  // Helper function to repopulate dropdown menu items
  const repopulateMenuItems = (menu, options, button) => {
    menu.textContent = "";
    options.forEach(item => {
      // Set button text to item text to find the minimum width to fit the content
      button.querySelector(".text").textContent = item
      menu.style.minWidth = `${button.offsetWidth}px`
      button.style.minWidth = `${menu.offsetWidth}px`

      const itemElement = document.createElement('li');
      itemElement.id = item;
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
  const selectItem = (item, button, menu, shouldUpdate=true) => {
    // Deselect all items
    menu.querySelectorAll("li").forEach(i => i.classList.remove("selected"));

    // Select current item
    item.classList.add("selected");
    button.querySelector(".text").textContent = item.textContent;

    // Close dropdown
    menu.classList.remove("active");

    if (menu.parentElement.id === "buildType") {
      const versionsDropdown = document.querySelector(".dropdown#version");
      const versionsButton = versionsDropdown.querySelector(".dropdown-button");
      const versionsMenu = versionsDropdown.querySelector(".dropdown-menu");
      repopulateMenuItems(versionsMenu, versions[item.textContent], versionsButton);
    }

    // Update links
    if (shouldUpdate) {
      setLinks()
    }
  };

  const getAvailableBuilds = () => {
    let buildTypes = ["release", "prerelease","nightly"];
    for (let i = 0; i < buildTypes.length; i++) {
      const type = buildTypes[i];
      if (type in versions) {
        if (versions[type].length !== 0) {
          return buildTypes.slice(i)
        }
      }
    }
  }

  // Fetch and populate dropdowns
  fetchOptions().then(data => {
    if (!data) return; // Exit if fetch failed

    versions = data.versions;
    // need to reverse to get latest as first element if appended in api
    for (type in versions) versions[type].reverse();
    options = data.options;
    commands = data.commands;

    const dropdowns = content.querySelectorAll(".dropdown");
    dropdowns.forEach(dropdown => {
      const button = dropdown.querySelector(".dropdown-button");
      const menu = dropdown.querySelector(".dropdown-menu");

      // Add dropdown toggle event
      button.addEventListener("click", () => menu.classList.toggle("active"));

      // Populate menu items
      let optionList;
      const availableBuilds = getAvailableBuilds();
      if (dropdown.id === "version") {
        optionList = versions[availableBuilds[0]];
      } else if (dropdown.id === "buildType") {
        optionList = availableBuilds;
      } else {
        optionList = options[dropdown.id];
      }
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
    setLinks(firstLoad=true)
  });
}
