import re, sys
from ruamel.yaml import YAML
from glob import glob


# CONSTANTS
FOLDER = './tokko-k8s/tokko-applications/ucp' 

# init yaml object
yaml = YAML()
yaml.preserve_quotes=True
yaml.indent = 2


def modify_files(function_name, new_tag):
    files = glob(FOLDER + '/**/*.yaml', recursive=True)
    files.extend(glob(FOLDER + '/**/*.yml', recursive=True))
    for file in files:
        change_file = False
        with open(file, "r") as stream:
            # read files and modify the objects
            k8s_objects = list(yaml.load_all(stream))
            for k8s_object in k8s_objects:
                print(k8s_object)
                # if the file uses krm function
                try:
                    if 'config.kubernetes.io/function' in k8s_object['metadata']['annotations']:  
                        image_name = k8s_object['metadata']['annotations']['config.kubernetes.io/function']
                        if function_name in image_name:  
                            change_file = True
                            new_image = re.sub(rf"{function_name}:.*", f"{function_name}:"+new_tag, image_name)
                            k8s_object['metadata']['annotations']['config.kubernetes.io/function'] = new_image
                except KeyError as e:
                    pass

        # write new objects to the files
        if change_file:
            with open(file, "w") as stream:
                yaml.dump_all(k8s_objects, stream)
