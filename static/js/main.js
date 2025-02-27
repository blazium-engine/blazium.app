Prism.plugins.NormalizeWhitespace.setDefaults({
  'remove-trailing': true,
  'remove-indent': true,
  'left-trim': true,
  'right-trim': false,
  'remove-initial-line-feed': true,
});
Prism.plugins.showInvisibles = {
  tab: '→',
  space: ' ',
};
// make prism treat js as gds
Prism.languages.javascript = Prism.languages.gdscript;
// Add the hook to ensure line numbers are always applied
Prism.hooks.add('before-sanity-check', function (env) {
  if (env.element && env.element.parentNode.tagName === 'PRE') {
    env.element.parentNode.classList.add('line-numbers');
  }
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
  hideHamMenu();
  checkNotice();

  // Trigger Prism
  content.querySelectorAll("pre > code").forEach((block) => {
    if (block.classList.contains("lang-")) {
      block.classList.add("language-none");
    } else {
      block.classList.add("language-gdscript");
    }
  });
  Prism.highlightAll();

  if (window.location.pathname.includes("/dev-tools/download")) {
    handleToolsDownload(content);
  } else if (window.location.pathname.includes("/download")) {
    handleEditorDownload(content);
  }
});

function acceptCookies() {
  const notice = document.getElementById("cookies-notice");
  if (notice) {
    notice.style.display = "none";
    const expiryDate = new Date();
    expiryDate.setFullYear(expiryDate.getFullYear() + 1); // Cookie expires in 1 year
    document.cookie = `cookiesAccepted=true;expires=${expiryDate.toUTCString()};path=/`;
  }
  checkNotice();
}

function deleteFirstPartyCookies() {
  document.cookie = `cookiesAccepted=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/`;
  alert("First-party cookies have been deleted.");
}

// Array to store removed iframes
var removedIframes = [];

// Function to remove all iframes
function removeAllIframes() {
  const iframes = document.querySelectorAll("iframe");
  removedIframes = []; // Clear previously stored iframes
  iframes.forEach(iframe => {
    removedIframes.push({
      element: iframe,
      parent: iframe.parentNode,
      nextSibling: iframe.nextSibling // To preserve the original position
    });
    iframe.remove(); // Remove the iframe from the DOM
  });
}

// Function to re-add removed iframes
function readdIframes() {
  removedIframes.forEach(({ element, parent, nextSibling }) => {
    if (nextSibling) {
      parent.insertBefore(element, nextSibling);
    } else {
      parent.appendChild(element);
    }
  });
  removedIframes = []; // Clear the array after re-adding
}

function dismissNotice() {
  const notice = document.getElementById("cookies-notice");
  if (notice) {
    notice.style.display = "none";
  }
}

function checkNotice() {
  const consent = document.cookie.split("; ").find(row => row.startsWith("cookiesAccepted="));
  const isAccepted = consent ? consent.split("=")[1] === "true" : false;
  if (!isAccepted) {
    removeAllIframes();
    document.querySelectorAll("div.iframe-placeholder").forEach((iframe) => {
      iframe.classList.remove("allow");
    });
    const notice = document.getElementById("cookies-notice");
    if (notice) {
      notice.style.display = "block";
    }
    return
  }
  if (removedIframes.length > 0) {
    readdIframes();
  }
  document.querySelectorAll("div.iframe-placeholder").forEach((iframe) => {
    iframe.classList.add("allow");
  });
}

function showHamMenu() {
  const menu = document.querySelector("#hamburger-nav")
  menu.style.display = "flex"
}

function hideHamMenu() {
  const menu = document.querySelector("#hamburger-nav")
  menu.style.display = "none"
}

function handleEditorDownload(content) {
  var versions;
  var options;
  var commands;

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

    // Hide arch and mono dropdown when android
    const androidToHide = content.querySelector("#no-android");
    if (androidToHide) {
      if (selectedOptions.os === "Android" || selectedOptions.os === "Horizon OS") {
        androidToHide.style.display = "none";
      } else if (androidToHide.style.display === "none") {
        androidToHide.style.display = "inline-block";
      }
    }

    // Update the links dynamically
    const sha256Button = content.querySelector("#sha256-btn");
    const sha512Button = content.querySelector("#sha512-btn");
    const changelogButton = content.querySelector("#changelog-btn");
    const templatesContainer = content.querySelector("#export-templates");
    const downloadButton = content.querySelector("#download-btn");
    const downloadCmd = content.querySelector("#download-cmd");

    const version = selectedOptions.version;
    const buildType = selectedOptions.buildType;

    if (sha256Button) {
      sha256Button.href = `/api/editor-sha/${buildType}/BlaziumEditor_v${version}.sha256`
    }
    if (sha512Button) {
      sha512Button.href = `/api/editor-sha/${buildType}/BlaziumEditor_v${version}.sha512`
    }

    if (changelogButton) {
      changelogButton.href = `/changelog?v=${buildType}_${version}`
    }

    if (templatesContainer) {
      const templates = templatesContainer.querySelector("#templates");
      const templatesMono = templatesContainer.querySelector("#templates-mono");

      const textContent = `${buildType} ${version}`

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

      const sha256Button = content.querySelector("#templates-sha256-btn");
      const sha512Button = content.querySelector("#templates-sha512-btn");
      if (sha256Button) {
        sha256Button.href = `/api/templates-sha/${buildType}/Blazium_v${version}_export_templates.sha256`
      }
      if (sha512Button) {
        sha512Button.href = `/api/templates-sha/${buildType}/Blazium_v${version}_export_templates.sha512`
      }
    }

    if (downloadButton) {
      const os = selectedOptions.os.toLowerCase();
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
  fetchOptions("/api/download-options/editor").then(data => {
    if (!data) return; // Exit if fetch failed

    versions = data.versions;
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
}

function getSystemInfo () {
  const userAgent = window.navigator.userAgent.toLowerCase();

  let os = undefined;
  if (/windows/.test(userAgent)) {
    os = "Windows";
  } else if (/mac/.test(userAgent)) {
    os = "MacOS";
  } else if (/android/.test(userAgent)) {
    os = "Android";
  } else if (/linux/.test(userAgent)) {
    os = "Linux";
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

// Helper function to fetch options
async function fetchOptions(url) {
  try {
    const response = await fetch(url, {
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

function handleToolsDownload(content) {
  var toolsVersions;
  var toolsNames;

  // Helper function to set links and text after selecting an option
  const setLinks = (firstLoad=false) => {
    if (firstLoad) {
      const {os,} = getSystemInfo();

      if (os) {
        const osDropdown = content.querySelector("#os");
        const osDropdownButton = osDropdown.querySelector(".dropdown-button");
        const osDropdownMenu = osDropdown.querySelector(".dropdown-menu");
        const itemToSelect = osDropdownMenu.querySelector("#"+os);

        selectItem(itemToSelect, osDropdownButton, osDropdownMenu, false);
      }

      const url = new URL(window.location.href);
      const tool = url.searchParams.get('tool');

      if (tool) {
        const toolDropdown = content.querySelector("#tool");
        const toolDropdownButton = toolDropdown.querySelector(".dropdown-button");
        const toolDropdownMenu = toolDropdown.querySelector(".dropdown-menu");
        const toolName = Object.keys(toolsNames).find(key => toolsNames[key] === tool);
        const itemToSelect = document.getElementById(toolName);

        selectItem(itemToSelect, toolDropdownButton, toolDropdownMenu, false);
      }
    }

    // Collect selected options from dropdowns
    const dropdowns = content.querySelectorAll(".dropdown");
    const selectedOptions = {};
    dropdowns.forEach(dropdown => {
      const selectedItem = dropdown.querySelector(".selected");
      selectedOptions[dropdown.id] = selectedItem.textContent;
    });

    // Update the links dynamically
    const downloadButton = content.querySelector("#download-btn");

    if (downloadButton) {
      const tool = selectedOptions.tool;
      const toolName = tool.toLowerCase().replace(" ", "-")
      const version = selectedOptions.version;
      var os = selectedOptions.os.toLowerCase();
      var isExe = ""
      
      if (os === "windows") {
        isExe = ".exe"
      }
      if (os === "macos") {
        os = "darwin"
      }

      downloadButton.href = `https://cdn.blazium.app/${toolsNames[tool]}/${os}/${version}/${toolName}${isExe}`

      const buttonLabel = downloadButton.querySelector("span");
      if (buttonLabel) {
        buttonLabel.textContent = `${tool} ${version}`;
      }
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

    if (menu.parentElement.id === "tool") {
      const versionsDropdown = document.querySelector(".dropdown#version");
      const versionsButton = versionsDropdown.querySelector(".dropdown-button");
      const versionsMenu = versionsDropdown.querySelector(".dropdown-menu");

      const codename = toolsNames[item.textContent]
      const versions = toolsVersions[codename]
      repopulateMenuItems(versionsMenu, versions, versionsButton);

      const url = new URL(window.location.href);
      url.searchParams.set("tool", codename);
      window.history.replaceState(null, "", url.toString());
    }

    // Update links
    if (shouldUpdate) {
      setLinks()
    }
  };

  // Fetch and populate dropdowns
  fetchOptions("/api/download-options/tools").then(data => {
    if (!data) return; // Exit if fetch failed

    toolsVersions = data.versions;
    toolsNames = data.names;

    const dropdowns = content.querySelectorAll(".dropdown");
    dropdowns.forEach(dropdown => {
      const button = dropdown.querySelector(".dropdown-button");
      const menu = dropdown.querySelector(".dropdown-menu");

      // Add dropdown toggle event
      button.addEventListener("click", () => menu.classList.toggle("active"));

      // Populate menu items
      let optionList;
      let names = Object.keys(toolsNames)
      if (dropdown.id === "tool") {
        optionList = names
      } else if (dropdown.id === "os") {
        optionList = data.os;
      } else {
        optionList = toolsVersions[toolsNames[names[0]]];
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
}