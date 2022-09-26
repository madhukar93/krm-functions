from yamlcompare import compareYaml

import unittest
import yaml
import sys

class TestYamlDiff(unittest.TestCase):

    def test_generated_objects(self):
        self.assertEqual(compareYaml(source_gen,dest_gen), False)

if __name__ == "__main__":
    file1 = sys.argv.pop()
    file2 = sys.argv.pop()

    with open(file1,'r') as rdr:
        source=rdr.read()
    
    with open(file2,'r') as rdr:
        dest=rdr.read()

    source_gen = yaml.load_all(source,Loader=yaml.FullLoader)
    dest_gen = yaml.load_all(dest,Loader=yaml.FullLoader)
    unittest.main()

