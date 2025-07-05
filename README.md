# git-delete-branch

A Git plugin to interactively delete local branches.

This tool provides an interactive command-line interface to view and delete multiple local Git branches.

## Features

- **Interactive UI:** Select branches to delete from a list in your terminal.
- **Multiple Selections:** Choose one or more branches to delete at once using a checkbox interface.
- **Incremental Search:** Filter branches by typing parts of the branch name.
- **Safe by Design:** Automatically excludes the currently checked-out branch from the deletion list.
- **Internationalization (i18n):** Automatically displays messages in English or Japanese based on your system's `LANG` environment variable.
- **Deletion Confirmation with Details:** Before deletion, review selected branches with their latest commit hash, author, date, and message.
- **Visual Merge Status:** Branches are visually marked as `(merged)` (green) or `(unmerged)` (red) in the selection list.

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) 1.16 or later must be installed.
- Git must be installed.
- [fzf](https://github.com/junegunn/fzf#installation) must be installed and available in your PATH.

### Steps

1.  **Clone or download this repository:**

    ```sh
    # Note: If you created the project locally, you can skip this step.
    git clone <repository_url>
    cd git-delete-branch
    ```

2.  **Build the executable:**

    Navigate to the project directory and run the following command to build the binary:

    ```sh
    go build
    ```

3.  **Set up a Git alias:**

    To use this tool like a native Git command (e.g., `git delete-branch`), add a Git alias. Open your global `.gitconfig` file or run the following command:

    ```sh
        ```sh
    git config --global alias.delete-branch '!$(pwd)/git-delete-branch'
    ```

    **Important:** This command assumes you are running it from the `git-delete-branch` project directory after building the executable.
    ```

    **Important:** Replace `/path/to/your/git-delete-branch/git-delete-branch` with the absolute path to the executable you built in the previous step.

## Usage

Run the tool using its Git alias:

```sh
git delete-branch
```

### Command-Line Options

- `-h`, `--help`: Show the help message.
- `-lang <lang>`: Specify the display language (`en` or `ja`). This overrides the system's `LANG` environment variable.

Example of specifying the language:

```sh
git delete-branch -lang ja
```

### How to Interact

1.  **Select Branches:**
    - The list of branches will show `(merged)` (green) or `(unmerged)` (red) next to each branch name to indicate its merge status with the current branch.
    - **Navigate:** Use the **Up/Down arrow keys** to move through the list of branches.
    - **Search:** Simply start typing to filter the list.
    - **Select:** Press the **Tab** key to select/deselect the highlighted branch (or **Shift+Tab** for multiple selections in some `fzf` configurations).
    - **Confirm Selection:** Press **Enter** to proceed to the confirmation step.

2.  **Confirm Deletion:**
    - After selecting branches, a summary of the chosen branches (including latest commit details) will be displayed.
    - A confirmation prompt will ask if you wish to proceed with the deletion.
    - Type `y` for Yes or `n` for No, then press **Enter**.
