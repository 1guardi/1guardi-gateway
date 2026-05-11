import yaml
from transformers import pipeline
import transformers

# Workaround for broken tokenizer_config.json in some models (e.g. patronus-studio/wolf-defender-prompt-injection)
# which mistakenly specify "tokenizer_class": "TokenizersBackend"
if not hasattr(transformers, "TokenizersBackend"):
    transformers.TokenizersBackend = transformers.PreTrainedTokenizerFast

def _build_pipeline(spec: dict):
    model = spec["model"]
    task = spec["task"]
    device = spec.get("device", -1)
    trust = spec.get("trust_remote_code", True)
    use_fast = spec.get("use_fast", True)

    try:
        from transformers import AutoTokenizer
        tokenizer = AutoTokenizer.from_pretrained(
            model,
            trust_remote_code=trust,
            use_fast=use_fast,
            extra_special_tokens={}  # Workaround for 'list' object has no attribute 'keys'
        )
    except Exception as e:
        print(f"[mlrunner] Failed to initialize tokenizer manually for {model}: {e}")
        tokenizer = None

    if tokenizer is not None:
        return pipeline(
            task,
            model=model,
            tokenizer=tokenizer,
            device=device,
            trust_remote_code=trust,
        )
    else:
        return pipeline(
            task,
            model=model,
            device=device,
            trust_remote_code=trust,
            use_fast=use_fast,
        )


def load_registry(config_path: str = "analyzers.yml") -> dict:
    """Load analyzer pipelines from YAML config. Called once at startup."""
    with open(config_path) as f:
        cfg = yaml.safe_load(f)

    registry = {}
    for name, spec in cfg.get("analyzers", {}).items():
        pipe = _build_pipeline(spec)
        registry[name] = {
            "pipeline": pipe,
            "max_length": spec.get("max_length", 512),
            "task": spec["task"],
        }
        print(f"[mlrunner] loaded analyzer: {name} ({spec['model']})")

    return registry
