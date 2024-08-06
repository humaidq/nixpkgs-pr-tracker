// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
import { Handle, Position } from "@xyflow/react";

function PRNode({ data }) {
  return (
    <>
      <div className={`pr-node ${data.Accepted ? "branch-accepted" : ""}`}>
        <Handle type="target" position={Position.Top} isConnectable={false} />
        <Handle
          type="source"
          position={Position.Bottom}
          isConnectable={false}
        />
        <label htmlFor="text">
          {data.Title} #
          <a href={`https://github.com/nixos/nixpkgs/pull/${data.ID}`}>
            {data.ID}
          </a>
        </label>
        <label>
          By{" "}
          <a href={`https://github.com/${data.AuthorUsername}`}>
            {data.AuthorUsername}
          </a>
        </label>
      </div>
    </>
  );
}

export default PRNode;
