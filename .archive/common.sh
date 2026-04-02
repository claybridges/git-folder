
echoError() {
    local red="$(tput bold; tput setaf 160)"
    echo -e "${red}ERROR: $1${reset}" >&2
}
