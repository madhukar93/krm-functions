

import dictdiffer
import logging as log

# return True if there are any differences in the objects

def compareYaml(source_gen, dest_gen):
    is_diff = False
    for source_data in source_gen:
        resource_kind = source_data['kind']
        for dest_data in dest_gen:
            if dest_data['kind'] == resource_kind:
                if source_data == dest_data:
                    continue
                else:
                    is_diff = True
                    differences = list(dictdiffer.diff(source_data, dest_data))
                    for diff in differences:
                        log.error(f"{resource_kind} - {diff} \n")
    return is_diff
            