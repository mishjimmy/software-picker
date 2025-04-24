# Paradigm Light Designer Launcher

A lightweight application for launching different versions of ETC Paradigm Light Designer installed on your Windows system.

## Features

- Automatically scans for Paradigm installations in the default location (Program Files (x86)\ETC)
- Lists all found Paradigm versions with a clean, modern UI
- Launch selected Paradigm Light Designer version with a single click
- Support for custom installation directories

## How It Works

The Paradigm Launcher scans for folders named "Paradigm X.X.X" (where X.X.X is the version number)
within the ETC directory. For each found version, it looks for the light_designer.exe within the
"Light Designer" subfolder.

## How to Use

1. **Launch the Application**:

   - The app will automatically scan the default ETC directory (Program Files (x86)\ETC)
   - All found Paradigm versions will appear as selectable options

2. **Select a Version**:

   - Click on a version from the list to select it
   - Version details will appear below the list

3. **Launch Light Designer**:

   - Click the "Launch Light Designer" button to start the selected version

4. **Custom Directory** (Optional):
   - If your Paradigm installations are in a non-standard location, enter the custom path
   - Click "Scan Directory" to search in the specified location

## Development

This application was built using:

- Go (backend)
- Wails (framework for Go desktop apps)
- Svelte (frontend UI)

### Building from Source

1. Install prerequisites:

   - Go (1.21+)
   - Wails v2
   - npm

2. Clone the repository and build:

   ```
   wails build
   ```

3. Find the executable in the `build/bin` directory

## License

MIT
