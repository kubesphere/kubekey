# NAME
**kk completion**: Generate shell completion scripts.

# DESCRIPTION
Generate shell completion scripts. Normally you don't need to do more extra work to have this feature if you've installed kk by brew.

# OPTIONS

## **--type**
Generate different types of shell.

# EXAMPLES
Installing bash completion on Linux.
If bash-completion is not installed on Linux, please install the 'bash-completion' package via your distribution's package manager.
Load the ks completion code for bash into the current shell.
```
source <(ks completion bash)
```
Write bash completion code to a file and source if from .bash_profile.
```
mkdir -p ~/.config/kk/ && kk completion --type bash > ~/.config/kk/completion.bash.inc
printf "
```
kk shell completion.
```
source '$HOME/.config/kk/completion.bash.inc'
" >> $HOME/.bash_profile
source $HOME/.bash_profile
```

In order to have good experience on zsh completion, ohmyzsh is a good choice.
Please install ohmyzsh by the following command.
```
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
```
Get more details about onmyzsh from: https://github.com/ohmyzsh/ohmyzsh

Load the kk completion code for zsh[1] into the current shell.
```
source <(kk completion --type zsh)
```
Set the kk completion code for zsh[1] to autoload on startup.
```
kk completion --type zsh > "${fpath[1]}/_kk"
```