# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


from typing import Optional, List, Dict, Any

from dataclasses import dataclass
from functools import lru_cache
from collections import defaultdict

import util

@dataclass(frozen=True)
class MIG:
  name: str
  target_size: int
  versions: List[str]
  region: str # Only regional MIGs are supported

  @classmethod
  def from_json(cls, jo: object) -> "MIG":
    return cls(
      name=jo["name"],
      target_size=jo["targetSize"],
      versions=[v["instanceTemplate"] for v in jo["versions"]],
      region=util.trim_self_link(jo["region"]),
    )

@lru_cache
def migs(lkp: util.Lookup, region: str) -> Dict[str, MIG]:
  resp = lkp.compute.regionInstanceGroupManagers().list(project=lkp.project, region=region).execute()
  return {m.name: m for m in [MIG.from_json(o) for o in resp["items"]]}


def create_mig(lkp: util.Lookup, mig: MIG) -> None:
  """
  Blocking creation of MIG
  """
  assert len(mig.versions) == 1

  op = lkp.compute.regionInstanceGroupManagers().insert(
    project=lkp.project,
    region=mig.region,
    body = dict(
      name=mig.name,
      versions=[dict(
        instanceTemplate=mig.versions[0])],
      targetSize=mig.target_size,
      # Sensible defaults, allow for changes when needed
      updatePolicy= { "instanceRedistributionType": "NONE" },
      instanceLifecyclePolicy= { "defaultActionOnFailure": "DO_NOTHING" },
      #distributionPolicy= {
      #  'zones': [
      #    {'zone': f'https://www.googleapis.com/compute/beta/projects/{project}/zones/{zone}'}],
      #},
    )
  ).execute()
  res = util.wait_for_operation(op) # !!!
  assert "error" not in res, f"{res}"


def _allocate_node_to_mig(lkp: util.Lookup, nodes: List[str]) -> Dict[str, List[str]]:
  def slice_id(node: str) -> int:
    return lkp.node_index(node) // 4 # !!!

  res : Dict[str, List[str]] = defaultdict(list)
  for _, nodes in util.groupby_unsorted(nodes, lkp.node_nodeset_name):
    nodes = list(nodes)
    ns = lkp.node_nodeset(nodes[0])
    for sid, nodes in util.groupby_unsorted(nodes, slice_id):
      mig_name = f"{lkp.cfg.slurm_cluster_name}-{ns.nodeset_name}-{sid}"
      res[mig_name] = list(nodes)
  return res

def resume_mig_nodes(lkp: util.Lookup, nodes: List[str]):
  for mig_name, nodes in _allocate_node_to_mig(lkp, nodes).items():
      _resume_mig_nodes(lkp, mig_name, nodes)


def _resume_mig_nodes(lkp: util.Lookup, mig_name: str, nodes: List[str]):
  assert nodes
  model = nodes[0]
  ns = lkp.node_nodeset(model)
  region = lkp.node_region(model)
  mig = migs(lkp, region).get(mig_name)

  if not mig:
    mig = MIG(
      name=mig_name,
      target_size=0,
      region=region,
      versions=[ns.instance_template])
    create_mig(lkp, mig)

  op = lkp.compute.regionInstanceGroupManagers().createInstances(
    project=lkp.project, region=mig.region, instanceGroupManager=mig.name,
    body=dict(
      instances=[{"name": node} for node in nodes]
  )).execute()

  res = util.wait_for_operation(op) # !!!
  assert "error" not in res, f"{res}"


def suspend_mig_nodes(lkp: util.Lookup, nodes: List[str]):
  for mig_name, nodes in _allocate_node_to_mig(lkp, nodes).items():
    _suspend_mig_nodes(lkp, mig_name, nodes)

  
def _suspend_mig_nodes(lkp: util.Lookup, mig_name: str, nodes: List[str]):
  assert nodes
  mig = migs(lkp, lkp.node_region(nodes[0])).get(mig_name)
  assert mig, f"{mig_name}"

  links = [
    f"zones/{inst.zone}/instances/{inst.name}"
    for inst in [
      lkp.instance(node) for node in nodes
    ]
  ]

  op = lkp.compute.regionInstanceGroupManagers().deleteInstances(
    project=lkp.project, region=mig.region, instanceGroupManager=mig.name,
    body=dict(
      instances=links,
      skipInstancesOnValidationError=True, # ??? !!!!
    )
  ).execute()

  res = util.wait_for_operation(op) # !!!
  assert "error" not in res, f"{res}"

  
def is_mig_node(node: str) -> bool:
  return "azlk" in node
  
