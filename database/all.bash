#!/usr/bin/env bash

set -e

origin_dir=$PWD
repository_root=$(git rev-parse --show-toplevel)
database_dir=${repository_root}/database
schema_dir=${database_dir}/schemas

# vet vets all databases using SQLC vet command. The function will create a database and apply the database schema
# before invoking the 'sqlc vet' command.
#
# The command receives empty string or a valid migration directory. In the case of empty string, it will loop through
# all directories.
#
# Usage: ./all.bash vet [empty/dir_name]
vet() {
	sqlc_exec $1 vet
}

# generate generates queries from database schema/migration by invoking 'sqlc generate command'. This command will automatically
# spawn postgresql container(in local) so we have a better result when generating sqlc codes(You can look more into the issue [here](https://github.com/sqlc-dev/sqlc/issues?q=is%3Aissue+label%3Aanalyzer)).
#
# Usage: ./all.bash generate [empty/dir_name]
generate() {
	sqlc_exec $1 generate
}

# excec is an execution wrapper to run sqlc commands.
sqlc_exec() {
	pgexec="psql -U postgres"
	# Since we are not in github actions, this means we don't have the services we need to be already up and running
	# so we need to invoke docker compose in our own machine to spawn the services.
	if [[ -z "${GITHUB_ACTIONS}" ]]; then
    	container_id=$(docker ps -aqf "name=sqlc-postgres")
    	pgexec="docker exec -it ${container_id} psql -U postgres"
	fi

	# Move to schema dir as we need to invoke the go run from the speciifc schema directory.
	cd $schema_dir
	if [[ "$1" == "" ]]; then
	   echo "directory parameter is needed"
	   exit 1;
	elif [[ -d ${1} ]]; then
		echo "dir=${1}"
	    # schema_dir is taken from the first directory from the whole generation path. For example, in case of go-example/ledger
		# we will take the go-example as the schema dir.
	    db_schema_dir=$(echo $1 | cut -d / -f1)
		# db_name is normalized by replacing '-' with '_' as hypen is not allowed in the database name.
	    db_name=$(echo $db_schema_dir | sed -e 's/-/_/')

		db_found=$(PGPASSWORD=postgres ${pgexec} -XtAc "SELECT 1 FROM pg_database WHERE datname='${db_name}'")
        if  [[ "${db_found}" != "" ]]; then
            echo "database ${db_name} already exists"
			cd $1
		else
			PGPASSWORD=postgres ${pgexec} -c "CREATE DATABASE ${db_name}"
			cd $1
			# Migrate up using golang-migrate CLI. In this case we will not specify the version of the migration because we want
			# to migrate up all migrations.
			migrate -verbose -source file://${schema_dir}/${db_schema_dir}/migrations -database postgres://postgres:postgres@localhost:5432/${db_name}?sslmode=disable up
        fi

		sqlc $2
		if [[ "$2" = "generate" ]]; then
			go run $origin_dir/main.go gengo . --sqlc_config=sqlc.yaml --db_schema_dir=${db_schema_dir} --db_name=${db_name}
		fi

        if  [[ "${db_found}" == "" ]]; then
			cd -
		fi
	fi
	# Move to before schema dir.
	cd -
}

# templategen creates the database directory if not exists and generate the sqlc configuration template.
# If the directory is already exists, then it will replace the sqlc coniguration in that directory with
# the new one.
#
# Usage: ./all.bash templategen [empty/dir_name]
templategen() {
	# If directory name is empty, then we should only loop through all dir and replace the template.
	if [[ "$1" == "" ]]; then
		for dir in */; do
			cp -f sqlc.yaml $dir/sqlc.yaml
			# Remove all '/' suffix.
			migration_dir=${dir%/*}
			sed -i '' "s/database_name/$migration_dir/g" $migration_dir/sqlc.yaml
		done
		exit 0;
	fi

	# Otherwise, create a new directory and generate new sqlc.yaml file.
	if [[ ! -d $1 ]]; then
		mkdir $1
	fi
	cp sqlc.yaml $1/sqlc.yaml
	sed -i '' "s/database_name/$1/g" $1/sqlc.yaml
}

# up craetes the database and apply schema for the database.
#
# Usage: ./all.bash up [dir_name]
up() {
    # If directory name is empty, then we should throw an error
	if [[ "$1" == "" ]]; then
	   echo "up: directory name is needed"
	   exit 1;
	fi

	pgexec="psql -U postgres"
	# Since we are not in github actions, this means we don't have the services we need to be already up and running
	# so we need to invoke docker compose in our own machine to spawn the services.
	if [[ -z "${GITHUB_ACTIONS}" ]]; then
		container_id=$(docker ps -aqf "name=sqlc-postgres")
		pgexec="docker exec -it ${container_id} psql -U postgres"
	fi

	# Move to the schema directory.
	cd $schema_dir
    # schema_dir is taken from the first directory from the whole generation path. For example, in case of go-example/ledger
    # we will take the go-example as the schema dir.
    db_schema_dir=$(echo $1 | cut -d / -f1)
    # db_name is normalized by replacing '-' with '_' as hypen is not allowed in the database name.
    db_name=$(echo $db_schema_dir | sed -e 's/-/_/')

    echo "creating and applying schema for database ${db_name}"

    db_found=$(PGPASSWORD=postgres ${pgexec} -XtAc "SELECT 1 FROM pg_database WHERE datname='${db_name}'")
    if  [[ "${db_found}" != "" ]]; then
        echo "database ${db_name} already exists"
        exit 0;
    fi

    PGPASSWORD=postgres ${pgexec} -c "CREATE DATABASE ${db_name}"
    cd $1
    # The mounted volume is in '/data' so we need to seek the schema there.
    PGPASSWORD=postgres ${pgexec} -d $db_name -f /data/$db_schema_dir/schema.sql
    cd -
    # Move to before schema dir
	cd -
}

# down teardown all databases and its schema.
#
# Usage: ./all.bash down
down() {
    docker compose -f ${database_dir}/docker-compose.yaml down --remove-orphans -v || true
}

case $1 in
	"vet")
		$1 $2
	;;
	"generate")
		$1 $2
	;;
	"templategen")
		$1 $2
	;;
	"up")
	   $1 $2
	;;
	"down")
	   $1
	;;
	*)
		echo "command $1 not supported"
		exit 1
	;;
esac
