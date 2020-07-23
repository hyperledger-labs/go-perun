#!/bin/bash

# Copyright 2020 - See NOTICE file for copyright holders.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
exit_code=0

# This script checks that all new commiters email adresses are contained in the NOTICE
# file.
# To check all commits on the current branch:   ./check-notice-authors.sh
# To check all commits newer than base:         ./check-notice-authors.sh base

# Call with an ancestor whereas all commits newer than the ancestor are checked.
base="$1"
if [ -z "$base" ]; then
    commits="$(git rev-list --reverse HEAD)"
else
    commits="$(git rev-list --reverse $base..HEAD)"
fi

# Authors found in commits and NOTICE.
declare -A known_authors
# Authors found only in commits but not NOTICE file.
declare -A assumed_authors

for c in $commits; do
    author=$(git show -s --format='%an <%ae>' $c)
    # Check Signed-Off-By message
    if ! git show -s --format='%B' $c | grep -wq "Signed-off-by: $author"; then
        echo "Commit $c is missing or has wrong 'Signed-off-by' message."
        exit_code=1
    fi

    # Get the notice file of the commit and check that the author is in it.
    notice=$(git show $c:NOTICE 2> /dev/null || true)
    for k in "${known_authors[@]}"; do
        a="${known_authors[$k]}"
        if ! echo "$notice" | grep -wq "$a"; then
            echo "Author '$a' was deleted from NOTICE in commit $c"
            exit_code=1
        fi
    done
    if [ -n "${assumed_authors[$author]}" ]; then
        continue
    fi
    # This must be the first commit of this author, since he should add himself
    # to the NOTICE file here.
    if ! echo "$notice" | grep -wq "$author"; then
        echo "Author '$author' is missing from NOTICE file and should have been added in commit $c."
        assumed_authors[$author]="$author"
        unset "known_authors[$author]"
        exit_code=1
    else
        known_authors[$author]="$author"
        unset "assumed_authors[$author]"
    fi
done

exit $exit_code
