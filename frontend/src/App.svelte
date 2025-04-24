<script lang="ts">
  import { onMount } from "svelte";
  import {
    ScanForParadigmVersions,
    LaunchParadigm,
    GetDefaultParadigmPath,
    BrowseDirectory,
    AddCustomParadigm,
    BrowserOpenURL,
  } from "../wailsjs/go/main/App.js";

  let paradigmVersions = [];
  let selectedVersion = null;
  let customDirectoryPath = "";
  let isLoading = false;
  let message = "";
  let messageTempId = null;
  let defaultEtcPath = "";

  // For file browser
  let showFileBrowser = false;
  let currentBrowsePath = "";
  let filesList = [];
  let breadcrumbs = [];

  // For menu
  let showFileMenu = false;

  // For theme
  let darkMode = true; // Default to dark mode

  let selectedVersionIndex = 0;
  /** @type {HTMLSelectElement} */
  let versionSelectRef;
  let isWindowFocused = true;

  // Focus trap function
  const focusTrap = (e) => {
    if (!isWindowFocused) {
      const focusWindow = () => {
        window.focus();
        document.body.focus();
        const versionSelect = document.getElementById("version-select");
        if (versionSelect) {
          versionSelect.focus();
          isWindowFocused = true;
        }
      };
      focusWindow();
    }
  };

  // Apply theme on component mount and when it changes
  $: if (darkMode !== undefined) {
    applyTheme();
  }

  onMount(async () => {
    try {
      // Load theme preference from localStorage
      const savedTheme = localStorage.getItem("paradigm-theme");
      if (savedTheme !== null) {
        darkMode = savedTheme === "dark";
      }

      // Apply the theme
      applyTheme();

      // Get default ETC path
      defaultEtcPath = await GetDefaultParadigmPath();

      if (defaultEtcPath) {
        // Auto-scan the default path on startup
        await scanForParadigm(defaultEtcPath);
      }

      // Add window focus event listeners
      window.addEventListener("focus", handleWindowFocus);
      window.addEventListener("blur", handleWindowBlur);

      // Listen for keyboard events from Go backend
      // @ts-ignore
      if (window.runtime?.EventsOn) {
        // @ts-ignore
        window.runtime.EventsOn("keydown", (key) => {
          // Create a synthetic event that matches the real keyboard event
          const event = {
            key,
            preventDefault: () => {},
            stopPropagation: () => {},
          };
          handleKeyPress(event);
        });
      }

      // Force initial focus
      const focusVersionSelect = () => {
        // @ts-ignore
        if (window.runtime?.WindowSetFocus) {
          // @ts-ignore
          window.runtime.WindowSetFocus();
        }
        window.focus();
        document.body.focus();
        const versionSelect = document.getElementById("version-select");
        if (versionSelect) {
          versionSelect.focus();
          isWindowFocused = true;
          // Add a visual indicator that the field is focused
          versionSelect.style.outline = "2px solid var(--accent-color)";
          versionSelect.style.outlineOffset = "2px";
        }
      };

      // Try to focus immediately
      focusVersionSelect();

      // Try again after a short delay
      setTimeout(focusVersionSelect, 100);
      setTimeout(focusVersionSelect, 500);
      setTimeout(focusVersionSelect, 1000);

      // Add window show event listener
      // @ts-ignore
      if (window.runtime?.EventsOn) {
        // @ts-ignore
        window.runtime.EventsOn("window:show", () => {
          focusVersionSelect();
        });
      }

      return () => {
        window.removeEventListener("focus", handleWindowFocus);
        window.removeEventListener("blur", handleWindowBlur);
        // @ts-ignore
        if (window.runtime?.EventsOff) {
          // @ts-ignore
          window.runtime.EventsOff("window:show");
          // @ts-ignore
          window.runtime.EventsOff("keydown");
        }
      };
    } catch (error) {
      showMessage("Error initializing: " + error);
    }
  });

  function applyTheme() {
    const root = document.documentElement;
    root.setAttribute("data-theme", darkMode ? "dark" : "light");
    root.style.backgroundColor = darkMode
      ? "rgba(0, 0, 0, 0.1)"
      : "rgba(255, 255, 255, 0.1)";
    root.style.color = darkMode ? "#f0f0f0" : "#333";
  }

  async function scanForParadigm(directoryPath) {
    message = "";
    isLoading = true;

    const pathToScan = directoryPath || defaultEtcPath || customDirectoryPath;

    try {
      if (!pathToScan) {
        throw new Error("No directory specified");
      }

      const result = await ScanForParadigmVersions(String(pathToScan));

      // Ensure we have a valid array, even if empty
      paradigmVersions = Array.isArray(result) ? result : [];

      if (paradigmVersions.length === 0) {
        showMessage("No Paradigm versions found in this directory.");
      } else {
        showMessage(
          `Found ${paradigmVersions.length} Paradigm versions!`,
          3000
        );
        // Auto-select the first version
        selectedVersion = paradigmVersions[0];
      }
    } catch (error) {
      paradigmVersions = []; // Reset to empty array on error
      showMessage("Error scanning for Paradigm: " + error);
    } finally {
      isLoading = false;
    }
  }

  async function openFileBrowser(initialPath) {
    try {
      currentBrowsePath = initialPath || defaultEtcPath || "C:\\";
      await refreshFileList();
      showFileBrowser = true;
      updateBreadcrumbs();
    } catch (error) {
      console.error("Error opening file browser:", error);
      showMessage("Error opening file browser: " + error);
    }
  }

  function toggleTheme() {
    darkMode = !darkMode;
    localStorage.setItem("paradigm-theme", darkMode ? "dark" : "light");
    applyTheme();
  }

  function updateBreadcrumbs() {
    // Split the path into individual components
    const parts = currentBrowsePath.split("\\");
    breadcrumbs = [];

    // Build up the breadcrumbs
    let currentPath = "";
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      if (part) {
        // For drive letters, add backslash
        if (i === 0 && part.includes(":")) {
          currentPath = part + "\\";
        } else {
          currentPath = currentPath ? currentPath + "\\" + part : part;
        }

        breadcrumbs.push({
          name: part,
          path: currentPath,
        });
      }
    }
  }

  async function refreshFileList() {
    try {
      filesList = await BrowseDirectory(currentBrowsePath);
      console.log("Files list:", filesList);
    } catch (error) {
      console.error("Error browsing directory:", error);
      showMessage("Error browsing directory: " + error);
    }
  }

  async function navigateTo(path) {
    currentBrowsePath = path;
    await refreshFileList();
    updateBreadcrumbs();
  }

  async function handleFileClick(filePath) {
    if (filePath.endsWith("/")) {
      // It's a directory, navigate to it
      await navigateTo(filePath.slice(0, -1));
    } else if (filePath.toLowerCase().endsWith(".exe")) {
      // It's an executable, select it
      await selectExecutable(filePath);
    }
  }

  async function selectExecutable(execPath) {
    try {
      const customVersion = await AddCustomParadigm(execPath);

      if (customVersion && customVersion.executablePath) {
        // Add this custom version to our list (at the beginning)
        paradigmVersions = [customVersion, ...paradigmVersions];

        // Select this version
        selectedVersion = customVersion;

        showMessage(`Selected executable: ${execPath}`, 2000);
        showFileBrowser = false;
      } else {
        showMessage("Invalid executable selected");
      }
    } catch (error) {
      console.error("Error selecting executable:", error);
      showMessage("Error selecting executable: " + error);
    }
  }

  async function launchParadigm(executablePath) {
    try {
      if (!executablePath) {
        throw new Error("No executable path provided");
      }

      const success = await LaunchParadigm(String(executablePath));
      if (success) {
        showMessage("Paradigm launched successfully!", 1500);
        // Close the launcher after successful launch
        setTimeout(() => {
          window.close();
          // For development mode, we need to close the window differently
          try {
            // @ts-ignore
            if (window.runtime?.Quit) {
              // @ts-ignore
              window.runtime.Quit();
            }
          } catch (e) {
            // Ignore any errors during quit
          }
        }, 500);
      } else {
        showMessage("Failed to launch Paradigm.");
      }
    } catch (error) {
      showMessage("Error launching Paradigm: " + error);
    }
  }

  function showMessage(msg, timeout = 0) {
    message = msg;

    // Clear previous timeout if exists
    if (messageTempId) {
      clearTimeout(messageTempId);
      messageTempId = null;
    }

    // Set new timeout if needed
    if (timeout > 0) {
      messageTempId = setTimeout(() => {
        message = "";
        messageTempId = null;
      }, timeout);
    }
  }

  function handleSelectVersion(version) {
    selectedVersion = version;
  }

  function handleKeyPress(event) {
    // Only handle keys if we have versions and window is focused
    if (!paradigmVersions.length || !isWindowFocused) return;

    // Prevent default behavior for our keys
    if (["ArrowUp", "ArrowDown", "Enter", "Escape"].includes(event.key)) {
      if (event.preventDefault) event.preventDefault();
      if (event.stopPropagation) event.stopPropagation();
    }

    switch (event.key) {
      case "ArrowUp":
        if (selectedVersionIndex > 0) {
          selectedVersionIndex--;
          selectedVersion = paradigmVersions[selectedVersionIndex];
          const select = document.getElementById("version-select");
          if (select instanceof HTMLSelectElement) {
            select.selectedIndex = selectedVersionIndex;
          }
        }
        break;
      case "ArrowDown":
        if (selectedVersionIndex < paradigmVersions.length - 1) {
          selectedVersionIndex++;
          selectedVersion = paradigmVersions[selectedVersionIndex];
          const select = document.getElementById("version-select");
          if (select instanceof HTMLSelectElement) {
            select.selectedIndex = selectedVersionIndex;
          }
        }
        break;
      case "Enter":
        if (selectedVersion && selectedVersion.executablePath) {
          launchParadigm(selectedVersion.executablePath);
        }
        break;
      case "Escape":
        // @ts-ignore
        if (window.runtime?.Quit) {
          // @ts-ignore
          window.runtime.Quit();
        }
        break;
    }
  }

  function getDisplayVersion(version) {
    return version && version.version ? version.version : "Unknown";
  }

  function getFilenameFromPath(path) {
    return path.split("\\").pop().replace("/", "");
  }

  function savePreferences() {
    if (customDirectoryPath) {
      scanForParadigm(customDirectoryPath);
    }
  }

  function toggleFileMenu() {
    showFileMenu = !showFileMenu;
  }

  function browseFiles() {
    openFileBrowser(defaultEtcPath);
  }

  function refreshVersions() {
    scanForParadigm(defaultEtcPath);
  }

  function handleWindowFocus() {
    isWindowFocused = true;
    // Focus the version select when window gains focus
    const versionSelect = document.getElementById("version-select");
    if (versionSelect) {
      // @ts-ignore
      if (window.runtime?.WindowSetFocus) {
        // @ts-ignore
        window.runtime.WindowSetFocus();
      }
      versionSelect.focus();
      versionSelect.style.outline = "2px solid var(--accent-color)";
      versionSelect.style.outlineOffset = "2px";
    }
  }

  function handleWindowBlur() {
    isWindowFocused = false;
    // Remove focus indicator when window loses focus
    const versionSelect = document.getElementById("version-select");
    if (versionSelect) {
      versionSelect.style.outline = "none";
    }
  }

  function handleTitleBarClick() {
    isWindowFocused = true;
    const versionSelect = document.getElementById("version-select");
    if (versionSelect) {
      versionSelect.focus();
    }
  }
</script>

<main
  class="app-container"
  on:focus={handleWindowFocus}
  on:blur={handleWindowBlur}
  on:click={handleWindowFocus}
  on:keydown={(e) => e.key === "Enter" && handleWindowFocus()}
  role="application"
>
  <div
    class="titlebar"
    on:click={handleTitleBarClick}
    on:keydown={(e) => e.key === "Enter" && handleTitleBarClick()}
  >
    <div class="titlebar-drag-region"></div>
  </div>
  <div
    class="container"
    on:click={() => window.focus()}
    on:keydown={(e) => e.key === "Enter" && window.focus()}
    role="none"
  >
    <div
      class="paradigm-container"
      on:click={() => window.focus()}
      on:keydown={(e) => e.key === "Enter" && window.focus()}
      role="none"
    >
      {#if isLoading}
        <div class="loading">
          <div class="spinner"></div>
          <div>Scanning for Paradigm versions...</div>
        </div>
      {:else if paradigmVersions.length > 0}
        <div
          class="version-panel"
          on:click={() => window.focus()}
          on:keydown={(e) => e.key === "Enter" && window.focus()}
          role="none"
        >
          <div class="version-controls">
            <div class="version-select-row">
              <label for="version-select">Version:</label>
              <select
                id="version-select"
                bind:value={selectedVersion}
                on:change={(e) => {
                  const target = e.target;
                  if (target instanceof HTMLSelectElement) {
                    selectedVersionIndex = target.selectedIndex;
                  }
                }}
                bind:this={versionSelectRef}
              >
                {#each paradigmVersions as version}
                  <option value={version}>
                    {getDisplayVersion(version)}
                  </option>
                {/each}
              </select>
            </div>
            <button
              class="launch-btn"
              on:click={() => launchParadigm(selectedVersion.executablePath)}
              disabled={!selectedVersion.executablePath}
            >
              Launch
              <span class="enter-icon">
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M9 10l-5 5 5 5" />
                  <path d="M20 4v7a4 4 0 01-4 4H4" />
                </svg>
              </span>
            </button>
          </div>

          {#if selectedVersion}
            <div
              class="path"
              on:click={() =>
                BrowserOpenURL(
                  `file:///${selectedVersion.path.replace(/\\/g, "/")}`
                )}
              on:keydown={(e) =>
                e.key === "Enter" &&
                BrowserOpenURL(
                  `file:///${selectedVersion.path.replace(/\\/g, "/")}`
                )}
              title="Open in File Explorer"
            >
              <code>{selectedVersion.path || "Unknown path"}</code>
            </div>
          {/if}

          {#if message}
            <div class="toast" class:fade-out={messageTempId !== null}>
              {message}
            </div>
          {/if}
        </div>
      {:else}
        <div class="empty-state">
          <p>No Paradigm versions found.</p>
          {#if defaultEtcPath}
            <p class="smaller">Scanned: {defaultEtcPath}</p>
          {/if}
          <p class="smaller">
            Make sure the directory contains folders named "Paradigm X.X.X"
          </p>
          <button
            class="browse-btn"
            on:click={() => openFileBrowser(defaultEtcPath)}
          >
            Browse Files
          </button>
        </div>
      {/if}
    </div>
  </div>

  {#if showFileBrowser}
    <div class="modal-overlay">
      <div class="file-browser">
        <div class="file-browser-header">
          <h3>Select Paradigm Executable</h3>
          <button class="close-btn" on:click={() => (showFileBrowser = false)}
            >×</button
          >
        </div>

        <div class="breadcrumb">
          {#each breadcrumbs as crumb, i}
            <button
              class="breadcrumb-item"
              on:click={() => navigateTo(crumb.path)}
              on:keydown={(e) => {
                if (e.key === "Enter") navigateTo(crumb.path);
              }}
              role="button"
            >
              {crumb.name}
            </button>
            {#if i < breadcrumbs.length - 1}
              <span class="separator">\</span>
            {/if}
          {/each}
        </div>

        <div class="file-list">
          {#if filesList.length === 0}
            <div class="empty-folder">This folder is empty</div>
          {:else}
            {#each filesList as file}
              <button
                class="file-item"
                class:folder={file.endsWith("/")}
                class:executable={file.toLowerCase().endsWith(".exe")}
                on:click={() => handleFileClick(file)}
                on:keydown={(e) => {
                  if (e.key === "Enter") handleFileClick(file);
                }}
                role="button"
              >
                <span class="file-icon"
                  >{file.endsWith("/")
                    ? "📁"
                    : file.toLowerCase().endsWith(".exe")
                      ? "⚙️"
                      : "📄"}</span
                >
                <span class="file-name">{getFilenameFromPath(file)}</span>
              </button>
            {/each}
          {/if}
        </div>

        <div class="file-browser-footer">
          <button on:click={() => (showFileBrowser = false)}>Cancel</button>
        </div>
      </div>
    </div>
  {/if}

  <div class="menu-bar" role="menubar">
    <button
      class="menu-item"
      role="menuitem"
      on:click={toggleFileMenu}
      on:keydown={(e) => e.key === "Enter" && toggleFileMenu()}
    >
      File
    </button>
    {#if showFileMenu}
      <div class="menu-dropdown" role="menu">
        <button
          class="menu-dropdown-item"
          role="menuitem"
          on:click={toggleTheme}
          on:keydown={(e) => e.key === "Enter" && toggleTheme()}
        >
          Toggle Theme
        </button>
        <div class="menu-separator"></div>
        <button
          class="menu-dropdown-item"
          role="menuitem"
          on:click={browseFiles}
          on:keydown={(e) => e.key === "Enter" && browseFiles()}
        >
          Browse Files
        </button>
        <button
          class="menu-dropdown-item"
          role="menuitem"
          on:click={refreshVersions}
          on:keydown={(e) => e.key === "Enter" && refreshVersions()}
        >
          Refresh
        </button>
      </div>
    {/if}
  </div>
</main>

<style>
  :root {
    /* Dark Mode Variables */
    --bg-color: rgba(0, 0, 0, 0.1);
    --text-color: #f0f0f0;
    --card-bg: rgba(0, 0, 0, 0.25);
    --card-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
    --menu-bg: rgba(0, 0, 0, 0.1);
    --menu-border: rgba(255, 255, 255, 0.15);
    --accent-color: #738adb;
    --launch-btn-bg: #2ecc71;
    --launch-btn-hover: #27ae60;
    --input-bg: rgba(0, 0, 0, 0.7);
    --input-border: rgba(255, 255, 255, 0.15);
    --modal-overlay: rgba(0, 0, 0, 0.8);
    --path-bg: rgba(0, 0, 0, 0.25);
    --separator-color: rgba(255, 255, 255, 0.15);
    --file-hover: rgba(255, 255, 255, 0.1);
    --folder-color: #5dadec;
    --executable-color: #2ecc71;
    --empty-color: rgba(255, 255, 255, 0.4);
    --spinner-border: rgba(255, 255, 255, 0.15);
    --message-bg: rgba(0, 0, 0, 0.25);
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
    margin: 0;
    padding: 0;
    box-sizing: border-box;
    background-color: transparent;
    color: var(--text-color);
    overflow: hidden;
  }

  :global(html) {
    background-color: var(--bg-color);
    backdrop-filter: blur(20px);
    overflow: hidden;
  }

  .app-container {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background-color: var(--bg-color);
    backdrop-filter: blur(20px);
    overflow: hidden;
    outline: none;
    transition: all 0.3s ease;
  }

  .app-container:not(:focus-within) {
    opacity: 0.7;
    background-color: rgba(0, 0, 0, 0.1);
  }

  .app-container:focus {
    outline: none;
  }

  .app-container:focus-within {
    box-shadow: 0 0 0 2px var(--accent-color);
  }

  .container {
    max-width: 500px;
    margin: 0 auto;
    padding: 1rem;
    flex: 1;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  h1 {
    text-align: center;
    color: var(--accent-color);
    margin-bottom: 1rem;
    font-size: 1.5rem;
  }

  h3 {
    color: var(--accent-color);
    margin: 0 0 0.8rem 0;
  }

  .paradigm-container {
    position: relative;
    background-color: var(--card-bg);
    border-radius: 8px;
    padding: 1.2rem;
    box-shadow: var(--card-shadow);
    min-height: 180px;
    backdrop-filter: blur(20px);
    overflow: hidden;
    width: 100%;
    max-width: 400px;
    border: 1px solid var(--menu-border);
  }

  .version-panel {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    overflow: hidden;
  }

  .version-controls {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    position: relative;
    padding: 4px 0;
  }

  .version-select-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .version-select-row label {
    font-weight: 600;
    font-size: 1.1rem;
    min-width: 60px;
    color: var(--text-color);
  }

  .version-select-row select {
    width: 150px;
    padding: 0.5rem;
    border-radius: 8px;
    border: 1px solid var(--input-border);
    background-color: var(--input-bg);
    color: var(--text-color);
    backdrop-filter: blur(20px);
    font-weight: 600;
    font-size: 1.1rem;
  }

  .version-select-row select:focus {
    outline: 2px solid var(--accent-color);
    outline-offset: 2px;
  }

  .launch-btn {
    padding: 0.75rem 2rem;
    font-size: 1.1rem;
    font-weight: 600;
    white-space: nowrap;
    background: linear-gradient(
      135deg,
      var(--launch-btn-bg),
      var(--launch-btn-hover)
    );
    color: white;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 120px;
    transition: all 0.2s ease;
    gap: 0.5rem;
    backdrop-filter: blur(10px);
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    position: relative;
    overflow: visible;
    margin: 4px 0;
  }

  .launch-btn::before {
    content: "";
    position: absolute;
    top: -4px;
    left: -4px;
    right: -4px;
    bottom: -4px;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.1),
      rgba(255, 255, 255, 0)
    );
    opacity: 0;
    transition: opacity 0.2s ease;
    border-radius: 10px;
    pointer-events: none;
  }

  .launch-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 12px rgba(0, 0, 0, 0.15);
  }

  .launch-btn:hover::before {
    opacity: 1;
  }

  .launch-btn:active {
    transform: translateY(0);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .launch-btn:disabled {
    background: linear-gradient(
      135deg,
      rgba(149, 165, 166, 0.5),
      rgba(149, 165, 166, 0.3)
    );
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  .enter-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 0.25rem;
    opacity: 0.9;
  }

  .enter-icon svg {
    width: 16px;
    height: 16px;
  }

  .path {
    font-family: "Consolas", "Monaco", "Courier New", monospace;
    background-color: var(--path-bg);
    padding: 0.8rem;
    border-radius: 6px;
    word-break: break-all;
    font-size: 0.85rem;
    line-height: 1.4;
    max-height: 60px;
    overflow: auto;
    backdrop-filter: blur(20px);
    border: 1px solid var(--input-border);
    margin-top: 0.5rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .path:hover {
    background-color: var(--file-hover);
    border-color: var(--accent-color);
  }

  .path code {
    display: block;
    white-space: pre-wrap;
    color: var(--text-color);
  }

  .toast {
    position: absolute;
    bottom: 1rem;
    left: 50%;
    transform: translateX(-50%);
    background-color: var(--message-bg);
    color: var(--text-color);
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    backdrop-filter: blur(20px);
    border: 1px solid var(--input-border);
    z-index: 100;
    animation: slide-up 0.3s ease-out;
  }

  .toast.fade-out {
    animation: fade-out 0.3s ease-out forwards;
  }

  @keyframes slide-up {
    from {
      transform: translate(-50%, 100%);
      opacity: 0;
    }
    to {
      transform: translate(-50%, 0);
      opacity: 1;
    }
  }

  @keyframes fade-out {
    from {
      opacity: 1;
    }
    to {
      opacity: 0;
    }
  }

  .browse-btn {
    background-color: var(--accent-color);
    font-size: 0.9rem;
    padding: 0.5rem 1rem;
  }

  .browse-btn:hover {
    opacity: 0.9;
  }

  .empty-state {
    text-align: center;
    color: var(--empty-color);
    padding: 2rem 0;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 0;
  }

  .spinner {
    border: 4px solid var(--spinner-border);
    border-left-color: var(--accent-color);
    border-radius: 50%;
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  .smaller {
    font-size: 0.8rem;
    margin: 0.4rem 0;
  }

  /* File browser styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--modal-overlay);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .file-browser {
    background-color: var(--card-bg);
    border-radius: 8px;
    width: 90%;
    max-width: 800px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    backdrop-filter: blur(20px);
  }

  .file-browser-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid var(--menu-border);
  }

  .file-browser-header h3 {
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: var(--empty-color);
    cursor: pointer;
    padding: 0.2rem 0.6rem;
  }

  .close-btn:hover {
    color: var(--text-color);
  }

  .breadcrumb {
    padding: 0.75rem 1rem;
    background-color: var(--message-bg);
    overflow-x: auto;
    white-space: nowrap;
    backdrop-filter: blur(20px);
  }

  .breadcrumb-item {
    color: var(--accent-color);
    cursor: pointer;
  }

  .breadcrumb-item:hover {
    text-decoration: underline;
  }

  .separator {
    margin: 0 0.3rem;
    color: var(--separator-color);
  }

  .file-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem 1rem;
    height: 300px;
  }

  .file-item {
    display: flex;
    align-items: center;
    padding: 0.6rem 0.5rem;
    cursor: pointer;
    border-radius: 4px;
  }

  .file-item:hover {
    background-color: var(--file-hover);
  }

  .file-icon {
    margin-right: 0.5rem;
    font-size: 1.2rem;
  }

  .file-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .folder {
    color: var(--folder-color);
  }

  .executable {
    color: var(--executable-color);
    font-weight: bold;
  }

  .empty-folder {
    text-align: center;
    padding: 2rem;
    color: var(--empty-color);
  }

  .file-browser-footer {
    padding: 1rem;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid var(--menu-border);
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .menu-bar,
  .menu-item,
  .menu-dropdown,
  .menu-dropdown-item,
  .menu-separator {
    display: none;
  }

  .titlebar {
    -webkit-app-region: drag;
    height: 32px;
    background: transparent;
    user-select: none;
    display: flex;
    justify-content: flex-end;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1000;
  }

  .titlebar-drag-region {
    flex: 1;
    height: 100%;
  }
</style>
