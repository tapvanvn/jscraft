#'
profile=""
if [ -f "$HOME/.bash_profile" ]; then
    profile=$HOME/.bash_profile
fi
if [ -f "$HOME/.zprofile" ]; then
    profile=$HOME/.zprofile
elif [ -f "$HOME/.zshrc" ]; then
    profile=$HOME/.zshrc
fi

sure_profile(){
    if ! [ -f "$profile" ];then 
        echo "we cannot detech profile file. please make sure you have .bash_profile if using bash, .zprofile or .zshrc if zsh"
        exit 1
    fi
}

sure_root(){
    if ! [ -d "$HOME/.newcontinent-team.com" ]; then 
        mkdir -p "$HOME/.newcontinent-team.com"
    fi
}

sure_root_script(){
    if ! [ -f "$HOME/.newcontinent-team.com/main.sh" ]; then 
        touch "$HOME/.newcontinent-team.com/main.sh"
    fi

    profile_content=$(<$profile)

    if ! [ "$profile_content" != "${profile_content/source ~\/.newcontinent-team.com\/main.sh/}" ]; then
        echo "\nsource ~/.newcontinent-team.com/main.sh\n" | tee -a $profile
    fi
}

sure_alter_to_root_script(){
    main_content=$(<$HOME/.newcontinent-team.com/main.sh)
    #find_content=$("source ")
    if ! [ "$main_content" != "${main_content/source $HOME\/.newcontinent-team.com\/jscraft\/jscraft.sh/}" ]; then
        printf "\nsource $HOME/.newcontinent-team.com/jscraft/jscraft.sh\n" | tee -a ~/.newcontinent-team.com/main.sh
    fi
}

check_go(){
    go_status="$(which go)"
    if [ -z $go_status ]; then
        echo "you need golang version 1.13.5 or later to use this bundle"
        exit 1
    fi
}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ "$#" -ne 0 ]; then
    action=$1
else 
    action="install"
fi

if [ $action = 'uninstall' ]; then 
    echo "uninstall"
else
    sure_profile
    check_go
    sure_root
    sure_root_script
    echo "check environment success\n\tprofile:$profile"
    mkdir -p $HOME/.newcontinent-team.com/jscraft
    cp -R $DIR/scripts/ $HOME/.newcontinent-team.com/jscraft/
    chmod -R +x  $HOME/.newcontinent-team.com

    cd $DIR/src 
    CGO_ENABLED=0 go build -o $HOME/.newcontinent-team.com/jscraft/jscraft
    echo "build jscraft success"
    sure_alter_to_root_script

    source $profile

fi

