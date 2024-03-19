import yaml

with open('example.yaml') as f:
    data = yaml.full_load(f)
    print(data)