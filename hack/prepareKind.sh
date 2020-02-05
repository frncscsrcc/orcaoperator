sudo kind delete cluster;
sudo kind create cluster;
kubectl apply -f resources/ignitor_crd.yaml;
kubectl apply -f resources/task_crd.yaml;
kubectl apply -f resources/task1.yaml;