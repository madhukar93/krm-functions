import re, sys
from ruamel.yaml import YAML
from glob import glob


FOLDER = './tokko-k8s/tokko-applications'

yaml = YAML()
yaml.preserve_quotes=True
yaml.indent = 2


def modify_files(function_name, new_tag):
    files = glob(FOLDER + '/**/*.yaml', recursive=True)
    files.extend(glob(FOLDER + '/**/*.yml', recursive=True))
    for file in files:
        change_file = False
        print("Handling file", file)
        with open(file, "r") as stream:
            try:
                k8s_objects = list(yaml.load_all(stream))
                for k8s_object in k8s_objects:
                    try:
                        print("apiVersion " + k8s_object["apiVersion"])
                        print("kind " + k8s_object["kind"])
                        print("metadata", k8s_object["metadata"], "\n--------------------")
                        image_name = k8s_object['metadata']['annotations']['config.kubernetes.io/function']
                        if function_name in image_name:  
                            change_file = True
                            new_image = re.sub(rf"{function_name}:.*", f"{function_name}:"+new_tag, image_name)
                            k8s_object['metadata']['annotations']['config.kubernetes.io/function'] = new_image
                    except KeyError:
                        continue
                    except TypeError:
                        continue
            except Exception:
                # catching too generic exception as try block only does load all objects from file.
                # Need to continue execution of the program for files that can be loaded and parsed.
                continue

        if change_file:
            with open(file, "w") as stream:
                yaml.dump_all(k8s_objects, stream)
