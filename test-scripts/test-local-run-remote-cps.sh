#! /usr/bin/env bash 

#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
#-----------------------------------------------------------------------------------------                   
#
# Objectives: Tests the cps rest layer as much as we can from a real testcase.
# 
#-----------------------------------------------------------------------------------------                   

if [[ "$CALLED_BY_MAIN" == "" ]]; then
    # Where is this script executing from ?
    BASEDIR=$(dirname "$0");pushd $BASEDIR 2>&1 >> /dev/null ;BASEDIR=$(pwd);popd 2>&1 >> /dev/null
    # echo "Running from directory ${BASEDIR}"
    export ORIGINAL_DIR=$(pwd)
    cd "${BASEDIR}"

    cd ${BASEDIR}/.. ; export PROJECT_DIR=$(pwd) ; cd - 2>&1 >> /dev/null

    export TEMP_DIR=${PROJECT_DIR}/temp/remote-cps

    #-----------------------------------------------------------------------------------------                   
    #
    # Set Colors
    #
    #-----------------------------------------------------------------------------------------                   
    bold=$(tput bold)
    underline=$(tput sgr 0 1)
    reset=$(tput sgr0)
    red=$(tput setaf 1)
    green=$(tput setaf 76)
    white=$(tput setaf 7)
    tan=$(tput setaf 202)
    blue=$(tput setaf 25)

    #-----------------------------------------------------------------------------------------                   
    #
    # Headers and Logging
    #
    #-----------------------------------------------------------------------------------------                   
    underline() { printf "${underline}${bold}%s${reset}\n" "$@" ;}
    h1() { printf "\n${underline}${bold}${blue}%s${reset}\n" "$@" ;}
    h2() { printf "\n${underline}${bold}${white}%s${reset}\n" "$@" ;}
    debug() { printf "${white}%s${reset}\n" "$@" ;}
    info() { printf "${white}➜ %s${reset}\n" "$@" ;}
    success() { printf "${green}✔ %s${reset}\n" "$@" ;}
    error() { printf "${red}✖ %s${reset}\n" "$@" ;}
    warn() { printf "${tan}➜ %s${reset}\n" "$@" ;}
    bold() { printf "${bold}%s${reset}\n" "$@" ;}
    note() { printf "\n${underline}${bold}${blue}Note:${reset} ${blue}%s${reset}\n" "$@" ;}


    #-----------------------------------------------------------------------------------------                   
    # Process parameters
    #-----------------------------------------------------------------------------------------                   
    GALASA_BOOTSTRAP=""

    while [ "$1" != "" ]; do
        case $1 in
            -b | --bootstrap )   shift
                                    GALASA_BOOTSTRAP=$1
                                    ;;
            -h | --help )           usage
                                    exit
                                    ;;
            * )                     error "Unexpected argument $1"
                                    usage
                                    exit 1
        esac
        shift
    done

    if [[ "${GALASA_BOOTSTRAP}" == "" ]]; then
        error "Need to use the --bootstrap parameter."
        usage
        exit 1  
    fi
else 
    export GALASA_BOOTSTRAP=${bootstrap}
    export PROJECT_DIR=${BASEDIR}
    export TEMP_DIR=${PROJECT_DIR}/temp/remote-cps
fi

#-----------------------------------------------------------------------------------------                   
# Functions
#-----------------------------------------------------------------------------------------                   
function usage {
    info "Syntax: test.sh [OPTIONS]"
    cat << EOF
Options are:
-b | --bootstrap : The url of the galasa api server. Mandatory. For example: https://my.server/api
-h | --help         : displays this help.
EOF
}



#-----------------------------------------------------------------------------------------                   
# More Functions...
#-----------------------------------------------------------------------------------------


#-----------------------------------------------------------------------------------------
function run_tests {
    h2 "Running the test code locally"
    # Add the "--log -" flag if you want to see more detailed output.
    
    ORIGINAL_BOOTSTRAP=$GALASA_BOOTSTRAP
    unset GALASA_BOOTSTRAP
    info "GALASA_BOOTSTRAP is not set. As we are launching a local run. Currently '$GALASA_BOOTSTRAP'"

    export GALASA_HOME=$TEMP_DIR/home
    info "GALASA_HOME is $GALASA_HOME"

    baseName="dev.galasa.cps.rest.test"
    cmd="${PROJECT_DIR}/bin/${binary} runs submit local --obr mvn:${baseName}/${baseName}.obr/0.0.1-SNAPSHOT/obr \
        --class ${baseName}.http/${baseName}.http.TestHttp \
        --bootstrap file://$TEMP_DIR/home/bootstrap.properties \
        --log -
       "
    info "Command is $cmd"
    $cmd
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to run the test code. Return code: ${rc}" ; exit 1 ; fi

    # Put back the original bootstrap variable.
    export GALASA_BOOTSTRAP=$ORIGINAL_BOOTSTRAP

    success "OK"
}

function build_galasa_home {
    h2 "Building galasa home"

    rm -fr $TEMP_DIR
    info "Creating temporary folder at $TEMP_DIR"
    mkdir $TEMP_DIR
    cd $TEMP_DIR

    export GALASA_HOME=$TEMP_DIR/home
    info "GALASA_HOME is $GALASA_HOME"


    cmd="${PROJECT_DIR}/bin/${binary} local init "
    info "Command is $cmd"
    $cmd
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to build galasa home. Return code: ${rc}" ; exit 1 ; fi


    galasaConfigStoreRestUrl=$(echo -n "${GALASA_BOOTSTRAP}" | sed "s/https:/galasacps:/g" | sed "s/\/bootstrap//g")

    cat << EOF >> $TEMP_DIR/home/bootstrap.properties

# These properties were added on the fly by the test script.

# Target the CPS on the ecosystem
framework.config.store=${galasaConfigStoreRestUrl}
framework.extra.bundles=dev.galasa.cps.rest
EOF

    success "OK"
}

function login_to_ecosystem {
    h2 "Logging into the ecosystem"
    info "GALASA_BOOTSTRAP is $GALASA_BOOTSTRAP"
    cmd="${PROJECT_DIR}/bin/${binary} auth login --log -"
    info "Command is $cmd"
    $cmd
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to login to the galasa server. Return code: ${rc}" ; exit 1 ; fi
    success "OK"
}

function logout_of_ecosystem {
    h2 "Logging out of the ecosystem"
    info "GALASA_BOOTSTRAP is $GALASA_BOOTSTRAP"
    cmd="${PROJECT_DIR}/bin/${binary} auth logout"
    info "Command is $cmd"
    $cmd
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to logout to the galasa server. Return code: ${rc}" ; exit 1 ; fi
    success "OK"
}

function generating_galasa_test_project {
    h2 "Generating galasa test project code..."
    cd $TEMP_DIR
    cmd="${PROJECT_DIR}/bin/${binary} project create --package dev.galasa.cps.rest.test --features http --obr --gradle --force --development "
    info "Command is $cmd"
    $cmd
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to generate galasa test project. Return code: ${rc}" ; exit 1 ; fi
    success "OK"
}


function build_test_project {
    h2 "Building the generated code..."
    cd $TEMP_DIR/dev.galasa.cps.rest.test
    gradle clean build publishToMavenLocal
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to build the generated test project code. Return code: ${rc}" ; exit 1 ; fi
    success "OK"
}

function log_variables {
    h2 "Logging variables"
    info "BASEDIR is $BASEDIR"
    info "PROJECT_DIR is $PROJECT_DIR"
    info "ORIGINAL_DIR is $ORIGINAL_DIR"
    info "TEMP_DIR is $TEMP_DIR"
    info "GALASA_BOOTSTRAP is $GALASA_BOOTSTRAP"
    info "Current folder is $(pwd)"
    success "OK"
}

function test_local_run_remote_cps() {
    h1 "Testing a local run, where the CPS draws properties from a remote ecosystem"
    cd $PROJECT_DIR
    log_variables
    build_galasa_home
    logout_of_ecosystem
    login_to_ecosystem
    generating_galasa_test_project
    build_test_project
    run_tests
    success "Local runs with remote CPS works."
}

if [[ "$CALLED_BY_MAIN" == "" ]]; then
    source $PROJECT_DIR/test-scripts/calculate-galasactl-executables.sh
    calculate_galasactl_executable
    test_local_run_remote_cps
fi
