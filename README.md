# git-delete-branch

A Git plugin to interactively delete local branches.

This tool provides an interactive command-line interface to view and delete multiple local Git branches.

## Features

- **Interactive UI:** Select branches to delete from a list in your terminal.
- **Multiple Selections:** Choose one or more branches to delete at once using a checkbox interface.
- **Incremental Search:** Filter branches by typing parts of the branch name.
- **Safe by Design:** Automatically excludes the currently checked-out branch from the deletion list.
- **Internationalization (i18n):** Automatically displays messages in English or Japanese based on your system's `LANG` environment variable.

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) 1.16 or later must be installed.
- Git must be installed.

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
    git config --global alias.delete-branch '!/path/to/your/git-delete-branch/git-delete-branch'
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

- **Navigate:** Use the **Up/Down arrow keys** to move through the list of branches.
- **Search:** Simply start typing to filter the list.
- **Select:** Press the **Spacebar** to select/deselect the highlighted branch.
- **Confirm:** Press **Enter** to delete all selected branches.
