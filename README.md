# git-rebuild-go
A Rebuild of Git's core functionality in Go.

## IMPORTANT: What this application can not do!
This application is not meant to be a full substitute for Git, and lacks many safety, error checking / handling, and common features found in Git.
A list of things this program can not do (as of when I created this) is given below:
+ Generate a tree from subdirectories (i.e. you can't have folders in your project)
  + See `mygit.go` line 115
+ Ignore files and folders listed in a .gitignore (.mygitignore?)
+ Create and move between branches
+ Checkout code from an earlier commit
+ Stash code
+ Staging: Everything is automatically staged, there is no such `add` command like there is in Git
+ Connect to remote repositories
  + No cloning
  + No pushing
  + No pulling
  + No fetching

This application is meant to replicate the core functionality of Git and learn how Git works under the hood by replicating Git's plumbing commands.

## How to Use
### Step 1. Make Sure you have Go Installed
In order to run this you need to have Go installed. To download and install Go, see the [documentation](https://go.dev/doc/install).

### Step 2. Run the Program
Once you have Go installed and have cloned this repository, to run the program, make sure you are in the same directory as the shell script and that the script has executable permissions, and type the following command to get started:
```
./mygit.sh
```
Running this will display a list of options to help you get started using the application.

## Git in 100 Seconds
The way Git works under the hood is fascinating. At its core, Git is a content-addressable filesystem, meaning it is a simple key-value data store. What this means is that you can insert any kind of content into a Git repository, for which Git will give you back a unique key you can use later to retrieve that content.

The way Git stores content is by using three different Git objects. They are
+ blobs: these snapshot the contents of a file and just the contents
+ trees: these snapshot groups of files, representing the folders and structure of your project
+ commits: these snapshot who, when, and why the content was stored

When given content, Git will generate the three different objects and put them in the `.git/objects` directory. Each object is stored as a single file, named as a 40 character SHA-1 checksum of the content. The subdirectories in `.git/objects` are named with the first two characters of the SHA-1, and the filename is the remaining 38 characters.

## Helpful Links
If you wish to read more about how Git works under the hood including Git's plumbing commands, see the [Git Objects Documentation](https://git-scm.com/book/en/v2/Git-Internals-Git-Objects). This was a valuable tool for this project and I highly recommend reading.

If you wish to look into Go, I will link the home page [here](https://go.dev/). As for Go's standard library, I found [this page](https://pkg.go.dev/std) to be another extremely valuable tool.
