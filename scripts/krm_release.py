import argparse
import os
import sys
from file_utils import modify_files
from git_utils import git_clone

arg_parser = argparse.ArgumentParser(description='Get the function name as input and update new release in the files accordingly')
arg_parser.add_argument('Function',
                       metavar='function',
                       type=str,
                       help='function name to update the new version tag')
arg_parser.add_argument('New_Tag',
                       metavar='new_version',
                       type=str,
                       help='new version tag to be updated in the files')

args = arg_parser.parse_args()

function_name = args.Function
new_tag = args.New_Tag

git_clone()
modify_files(function_name=function_name, new_tag=new_tag)
