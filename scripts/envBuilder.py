import os
import venv
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("envDir",help="The dir path to create")
builder=venv.EnvBuilder(clear=True,with_pip=True)
args = parser.parse_args()
envDir=args.envDir
builder.create(envDir)