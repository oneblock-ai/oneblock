# Model Management and Serving

## Summary
Deploy a custom LLM on the 1Block.AI platform is a key component for in-house machine learning works. This enhancement is to provide a way to allow users create, track and mange version of models, and to deploy the model and business logic as a service.

### Related Issues

https://github.com/oneblock-ai/oneblock/issues/25

## Motivation
- ModelTemplate provides a way to allow users to create, define, and manage LLMs with different configurations and versions.
- MLServe provides a way to allow users to deploy machine learning models and business logic as a service to a ML cluster(RayCluster). 
  - It is a scalable model serving tool for building online inference APIs.
  - It is a single toolkit to serve everything from deep learning models built with frameworks like PyTorch, TensorFlow, and Keras, to Scikit-Learn models, to arbitrary Python business logic.
  - It has several features and performance optimizations for serving Large Language Models such as response streaming, dynamic request batching, multi-node/multi-GPU serving, etc.

### Goals
- Provide a list of built-in OSS models, e.g., mistral-7b, llama-7b, falcon-7b, gemma-7b, etc.
- Allow users to create, track and manage versions of models.
- Allow one-click deployment using a model template.
- Provide a way to allow users to deploy ML models and business logic as a service.

### Non-goals [optional]

N/A

## Proposal

### User Stories

#### Story 1
As a user, I want to deploy a pre-configured LLM template as an inference service with one-click deployment.

#### Story 2
As a user, I want to create a new large language model by providing a hugging face model ID(e.g., google/gemma-7b).

### User Experience In Detail

#### Story 1
As a user, I want to deploy the built-in LLM template as an inference service API with one-click .
1. Select a built-in LLM template from the model template listing page.
2. Click the "Deploy" button to deploy the model as a service.
3. Fill in the serve form, e.g., name, namespace, MLCluster, etc.
4. Click the "Save" button to deploy the model as an inference service API.
```YAML
apiVersion: ml.oneblock.ai/v1alpha1
kind: MLService
metadata:
  name: mistral-7b-llm
  namespace: default
spec:
  modelTemplateVersionRef:
    name: mistral-7b-instruct-v02
    namespace: default
  mlClusterRef:
    name: mistral-7b-cluster
    namespace: default
    rayClusterSpec:
      version: 0.5.0
      image: "anyscale/ray-llm"
      headGroupSpec:
        rayStartParams:
          dashboard-host: "0.0.0.0"
          num-cpus: "0"
        volume:
          name: mistral-7b-llm-head-log
          spec:
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 5Gi
      workerGroupSpec:
      - name: small-wg
        replicas: 1
        minReplicas: 1
        maxReplicas: 5
        acceleratorTypes:
          Nvidia-A10: 1
        rayStartParams:
          block: 'true'
        resources:
          limits:
            cpu: 8
            memory: 12Gi
          requests:
            cpu: 4
            memory: 8Gi
        nodeSelector:
          nvidia.com/product-type: nvidia-A10
        volume:
          name: mistral-7b-llm-small-wg-log
          spec:
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 5Gi
```

#### Story 2
As a user, I want to create a new large language model by providing a hugging face model ID(e.g., google/gemma-7b).
1. Go to the model template page.
2. Click the "Create" button to create a new model template.
3. Fill in the model template form, e.g., name, namespace, hf_model_id, max_total_tokens, engine_kwargs, etc.
4. Click the "Save" button to create a new model template.

Example of ModelTemplate YAML:
```YAML
apiVersion: ml.oneblock.ai/v1alpha1
kind: ModelTemplate
metadata:
  name: mistral-7b-instruct
  namespace: default
spec:
  description: An open-source mistral-7b-instruct model.
---
apiVersion: ml.oneblock.ai/v1alpha1
kind: ModelTemplateVersion
metadata:
  name: mistral-7b-instruct-v02
  namespace: default
spec:
  engine_config:
    # Model id - this is a RayLLM id
    model_id: "mistralai/Mistral-7B-Instruct-v0.2"
    # Id of the model on Hugging Face Hub. Can also be a disk path. Defaults to model_id if not specified.
    hf_model_id: "mistralai/Mistral-7B-Instruct-v0.2"
    type: VLLMEngine # TRTLLMEngine VLLMEngine
    # LLM engine keyword arguments passed when constructing the model.
    vLLMArgs: |
      trust_remote_code: true
      max_num_batched_tokens: 2048
      max_num_seq: 32
      dtype: auto
  deploymentConfig:
    replicas: 1
    maxReplicas: 5
    maxConcurrentQueries: 32
  scalingConfig:
    # If using multiple GPUs set num_gpus_per_worker to be 1 and then set num_workers to be the number of GPUs you want to use.
    numWorkers: 1
    numGPUsPerWorker: 1
    numCPUsPerWorker: 3
  workerGroupSpec:
    name: default-worker
    resources:
      limits:
        cpu: "8"
        memory: "12Gi"
        nvidia.com/gpu: "1"
      requests:
        cpu: "4"
        memory: "8Gi"
        nvidia.com/gpu: "1"
```

### API changes
Three new CRDs will be added to the API:
- `ModelTemplate`: A model template is a blueprint for creating a new model.
	- `ModelTemplateVersion`: Specify the details of the model template with an auto-generated version number.
- `MLService`: A MLServe is a model serving configuration for deploying a model as a service.

## Design

### Pre-requisite
- Worker node with GPU accelerator is required for LLM serving.

### Implementation Overview

- Users can define the ModelTemplate by following the [RayServe](https://docs.ray.io/en/latest/serve/index.html) and [vLLM](https://docs.vllm.ai/en/latest/index.html) config spec.
  - ModelTemplateVersion spec is not editable, it only support modification by creating a new ModelTemplateVersion.
- ModelTemplateVersion is served using MLService CRD, it will auto-create RayService and RayCluster for a generative AI infrastructure management.

### Test plan
User can test the model serving by using the following steps:
1. Create or select an existing model template.
2. Deploy the model as a service.
3. Wait for the serve to be ready.
4. Test the model inference API using the OpenAI python SDK. e.g.,
```python
import openai

query = "Once upon a time,"

client = openai.OpenAI(
    base_url = "https://public-ml-cluster.oneblock-public:8000/v1",
    api_key = "not_a_real_api_key"
)
# Note: not all arguments are currently supported and will be ignored by the backend.
chat_completion = client.chat.completions.create(
    model="mistralai/Mistral-7B-Instruct-v0.2",
    messages=[{"role": "system", "content": "You are a helpful assistant."}, 
              {"role": "user", "content": query}],
    temperature=0.1,
)
print(chat_completion.choices[0].message.content)
```

### Upgrade strategy

N/A

## Note [optional]

Additional nodes.
