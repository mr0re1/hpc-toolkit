## Local Docker
```sh
# Running from the repo root
docker build -t ghpc_dock - < tools/cloud-build/public/Dockerfile
docker run \
    -v $(pwd):/ghpc_dir \
    ghpc_dock \
    create /ghpc_dir/examples/hpc-slurm.yaml -o /ghpc_dir \
    -l IGNORE --vars=project_id=io-playground
```

## CloudBuild
```sh
# Build container image
gcloud builds submit \
    -t us-central1-docker.pkg.dev/io-playground/kartinki/ghpc_dock \
    tools/cloud-build/public
```

## Benchmark
2 steps (gcr.io/cloud-builders/gsutil -m cp -r) - 30 sec
