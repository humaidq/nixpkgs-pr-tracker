// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
import { Handle, Position } from "@xyflow/react";

function BranchNode({ data }) {
  return (
    <>
      <div className={`branch-node ${data.Accepted ? "branch-accepted" : ""}`}>
        <Handle type="target" position={Position.Top} isConnectable={false} />
        <Handle
          type="source"
          position={Position.Bottom}
          isConnectable={false}
        />
        <a href={data.HydraLink} target="_blank">
          <label>{data.BranchName}</label>
        </a>
      </div>
    </>
  );
}

export default BranchNode;
