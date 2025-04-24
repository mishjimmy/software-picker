<script lang="ts">
  import { onMount } from "svelte";
  import {
    ScanForParadigmVersions,
    LaunchParadigm,
    GetDefaultParadigmPath,
    BrowseDirectory,
    AddCustomParadigm,
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

      // Focus the window immediately
      window.focus();

      // Add keyboard event listener to the window with capture phase
      window.addEventListener("keydown", handleKeyPress, true);

      // Get default ETC path
      defaultEtcPath = await GetDefaultParadigmPath();

      if (defaultEtcPath) {
        // Auto-scan the default path on startup
        await scanForParadigm(defaultEtcPath);
      }

      // Focus the version select after versions are loaded and DOM is rendered
      setTimeout(() => {
        if (paradigmVersions.length > 0) {
          const versionSelect = document.getElementById("version-select");
          if (versionSelect) {
            versionSelect.focus();
          }
        }
      }, 100);
    } catch (error) {
      showMessage("Error initializing: " + error);
    }

    return () => {
      // Cleanup keyboard event listener
      window.removeEventListener("keydown", handleKeyPress, true);
    };
  });

  function applyTheme() {
    const root = document.documentElement;
    root.setAttribute("data-theme", darkMode ? "dark" : "light");
    root.style.backgroundColor = darkMode
      ? "rgba(0, 0, 0, 0.05)"
      : "rgba(255, 255, 255, 0.05)";
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
          2000
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
        showMessage("Paradigm launched successfully!", 500);
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
    // Only handle these keys if we have versions loaded
    if (paradigmVersions.length === 0) {
      return;
    }

    // Prevent default behavior for our keys
    if (["Enter", "ArrowUp", "ArrowDown"].includes(event.key)) {
      event.preventDefault();
      event.stopPropagation();
    }

    if (event.key === "Enter") {
      if (selectedVersion && selectedVersion.executablePath) {
        launchParadigm(selectedVersion.executablePath);
      }
    } else if (event.key === "ArrowUp") {
      const currentIndex = paradigmVersions.indexOf(selectedVersion);
      if (currentIndex > 0) {
        selectedVersion = paradigmVersions[currentIndex - 1];
      }
    } else if (event.key === "ArrowDown") {
      const currentIndex = paradigmVersions.indexOf(selectedVersion);
      if (currentIndex < paradigmVersions.length - 1) {
        selectedVersion = paradigmVersions[currentIndex + 1];
      }
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

  // Remove window focus/blur handlers since we don't need the alerts
  function handleWindowFocus() {
    window.focus();
  }

  function handleWindowBlur() {
    window.focus();
  }
</script>

<main
  class="app-container"
  on:focus={handleWindowFocus}
  on:blur={handleWindowBlur}
  on:click={() => {
    window.focus();
    handleWindowFocus();
    // Focus the version select if it exists
    const versionSelect = document.getElementById("version-select");
    if (versionSelect) {
      versionSelect.focus();
    }
  }}
  on:keydown={handleKeyPress}
  role="application"
  tabindex="-1"
>
  <div
    class="container"
    on:click={() => window.focus()}
    on:keydown={(e) => e.key === "Enter" && window.focus()}
    role="none"
    tabindex="-1"
  >
    <div
      class="paradigm-container"
      on:click={() => window.focus()}
      on:keydown={(e) => e.key === "Enter" && window.focus()}
      role="none"
      tabindex="-1"
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
          tabindex="-1"
        >
          <div class="version-controls">
            <div class="version-select-row">
              <label for="version-select">Version:</label>
              <select id="version-select" bind:value={selectedVersion}>
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
              <span class="enter-icon">↵</span>
            </button>
          </div>

          {#if selectedVersion}
            <div class="version-details">
              <div class="path">
                {selectedVersion.path || "Unknown path"}
              </div>
            </div>
          {/if}

          {#if message}
            <div class="message">{message}</div>
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
            <span
              class="breadcrumb-item"
              on:click={() => navigateTo(crumb.path)}
              on:keydown={(e) => {
                if (e.key === "Enter") navigateTo(crumb.path);
              }}
              tabindex="0"
              role="button"
            >
              {crumb.name}
            </span>
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
              <div
                class="file-item"
                class:folder={file.endsWith("/")}
                class:executable={file.toLowerCase().endsWith(".exe")}
                on:click={() => handleFileClick(file)}
                on:keydown={(e) => {
                  if (e.key === "Enter") handleFileClick(file);
                }}
                tabindex="0"
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
              </div>
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
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      on:click={toggleFileMenu}
      on:keydown={(e) => e.key === "Enter" && toggleFileMenu()}
    >
      File
    </div>
    {#if showFileMenu}
      <div class="menu-dropdown" role="menu">
        <div
          class="menu-dropdown-item"
          role="menuitem"
          tabindex="0"
          on:click={toggleTheme}
          on:keydown={(e) => e.key === "Enter" && toggleTheme()}
        >
          Toggle Theme
        </div>
        <div class="menu-separator"></div>
        <div
          class="menu-dropdown-item"
          role="menuitem"
          tabindex="0"
          on:click={browseFiles}
          on:keydown={(e) => e.key === "Enter" && browseFiles()}
        >
          Browse Files
        </div>
        <div
          class="menu-dropdown-item"
          role="menuitem"
          tabindex="0"
          on:click={refreshVersions}
          on:keydown={(e) => e.key === "Enter" && refreshVersions()}
        >
          Refresh
        </div>
      </div>
    {/if}
  </div>
</main>

<style>
  :root {
    /* Light Mode Variables */
    --bg-color: rgba(255, 255, 255, 0.05);
    --text-color: #333;
    --card-bg: rgba(255, 255, 255, 0.1);
    --card-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
    --menu-bg: rgba(255, 255, 255, 0.05);
    --menu-border: rgba(255, 255, 255, 0.1);
    --accent-color: #3b3b8c;
    --launch-btn-bg: #27ae60;
    --launch-btn-hover: #219653;
    --input-bg: rgba(255, 255, 255, 0.1);
    --input-border: rgba(255, 255, 255, 0.2);
    --modal-overlay: rgba(0, 0, 0, 0.5);
    --path-bg: rgba(255, 255, 255, 0.05);
    --separator-color: rgba(0, 0, 0, 0.1);
    --file-hover: rgba(255, 255, 255, 0.1);
    --folder-color: #3498db;
    --executable-color: #27ae60;
    --empty-color: rgba(0, 0, 0, 0.3);
    --spinner-border: rgba(0, 0, 0, 0.1);
    --message-bg: rgba(255, 255, 255, 0.1);
  }

  [data-theme="dark"] {
    /* Dark Mode Variables */
    --bg-color: rgba(0, 0, 0, 0.05);
    --text-color: #f0f0f0;
    --card-bg: rgba(0, 0, 0, 0.1);
    --card-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
    --menu-bg: rgba(0, 0, 0, 0.05);
    --menu-border: rgba(255, 255, 255, 0.05);
    --accent-color: #738adb;
    --launch-btn-bg: #2ecc71;
    --launch-btn-hover: #27ae60;
    --input-bg: rgba(0, 0, 0, 0.1);
    --input-border: rgba(255, 255, 255, 0.05);
    --modal-overlay: rgba(0, 0, 0, 0.7);
    --path-bg: rgba(0, 0, 0, 0.1);
    --separator-color: rgba(255, 255, 255, 0.1);
    --file-hover: rgba(255, 255, 255, 0.05);
    --folder-color: #5dadec;
    --executable-color: #2ecc71;
    --empty-color: rgba(255, 255, 255, 0.3);
    --spinner-border: rgba(255, 255, 255, 0.1);
    --message-bg: rgba(0, 0, 0, 0.1);
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
    background-color: var(--card-bg);
    border-radius: 8px;
    padding: 1.2rem;
    box-shadow: var(--card-shadow);
    min-height: 180px;
    backdrop-filter: blur(20px);
    overflow: hidden;
    width: 100%;
    max-width: 400px;
  }

  .version-panel {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    overflow: hidden;
  }

  .version-controls {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .version-select-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .version-select-row label {
    font-weight: 500;
    min-width: 60px;
  }

  .version-select-row select {
    width: 150px;
    padding: 0.5rem;
    border-radius: 8px;
    border: 1px solid var(--input-border);
    background-color: var(--input-bg);
    color: var(--text-color);
    backdrop-filter: blur(20px);
  }

  .launch-btn {
    padding: 0.5rem 1.5rem;
    font-size: 1rem;
    white-space: nowrap;
    background-color: rgba(39, 174, 96, 0.8);
    color: white;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 100px;
    transition: all 0.2s;
    gap: 0.5rem;
    backdrop-filter: blur(10px);
  }

  .launch-btn:hover {
    background-color: rgba(33, 150, 83, 0.9);
  }

  .launch-btn:disabled {
    background-color: rgba(149, 165, 166, 0.5);
    cursor: not-allowed;
  }

  .enter-icon {
    font-size: 0.9em;
    opacity: 0.8;
  }

  .version-details {
    background-color: var(--message-bg);
    border-radius: 8px;
    padding: 0.8rem;
    backdrop-filter: blur(20px);
  }

  .path {
    font-family: monospace;
    background-color: var(--path-bg);
    padding: 0.5rem;
    border-radius: 4px;
    word-break: break-all;
    font-size: 0.8rem;
    max-height: 40px;
    overflow: hidden;
    backdrop-filter: blur(20px);
  }

  .message {
    margin-top: 1rem;
    padding: 0.75rem;
    border-radius: 4px;
    background-color: var(--message-bg);
    border-left: 4px solid var(--accent-color);
    backdrop-filter: blur(20px);
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

  .version-select-row select:focus {
    outline: 2px solid var(--accent-color);
    outline-offset: 2px;
  }
</style>
