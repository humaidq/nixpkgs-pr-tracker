// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
import { useCallback } from "react";
import { Handle, Position } from "@xyflow/react";

function PRNode({ data }) {
  return (
    <>
      <div className={"pr-node"}>
        <Handle type="target" position={Position.Top} isConnectable={false} />
        <Handle
          type="source"
          position={Position.Bottom}
          isConnectable={false}
        />
        <label htmlFor="text">
          {data.Title} #{data.ID}
        </label>
      </div>
    </>
  );
}

export default PRNode;
