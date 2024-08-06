// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
import { useState } from "react";
import { ReactFlow, Background, Controls, MiniMap, Panel } from "@xyflow/react";
import "./App.css";
import BranchNode from "./components/BranchNode.jsx";
import PRNode from "./components/PRNode.jsx";
import "@xyflow/react/dist/style.css";
import dagre from "dagre";
import { toast, ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";

const flattenTree = (pr) => {
  const prId = pr["ID"];
  const nodes = [
    {
      id: `${prId}`,
      position: { x: 0, y: 0 },
      data: pr,
      type: "pr",
    },
  ];
  const edges = [];

  if (pr.Branches != null) {
    const head = pr.Branches;
    var style = {
      strokeWidth: 3,
    };
    if (head.Accepted) {
      style.stroke = "#226b23";
    } else {
      style.stroke = "#a49003";
    }
    edges.push({
      id: `${prId}-root`,
      source: `${prId}`,
      target: `${prId}-${head.BranchName}`,
      animated: true,
      style: style,
    });
  }

  const traverse = (node, parent = null) => {
    if (!node) return;

    // Add the current node to the list of nodes
    nodes.push({
      id: `${prId}-${node.BranchName}`,
      position: { x: 0, y: 0 },
      data: node,
      type: "branch",
    });

    // If there's a parent, add an edge from parent to this node
    if (parent) {
      var style = {
        strokeWidth: 3,
      };
      if (parent.Accepted && node.Accepted) {
        style.stroke = "#226b23";
      } else if (parent.Accepted && !node.Accepted) {
        style.stroke = "#a49003";
      }
      edges.push({
        id: `${prId}-${parent.BranchName}-${node.BranchName}`,
        source: `${prId}-${parent.BranchName}`,
        target: `${prId}-${node.BranchName}`,
        animated: true,
        style: style,
      });
    }

    // If the node has children, traverse them
    if (node.Children) {
      node.Children.forEach((child) => traverse(child, node));
    }
  };

  traverse(pr["Branches"]);
  console.log(nodes);
  console.log(edges);
  return { nodes, edges };
};

const nodeTypeSizes = {
  branch: { width: 150, height: 40 },
  pr: { width: 300, height: 100 },
};

const getLayoutedElements = (nodes, edges, direction = "TB") => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  const isHorizontal = direction === "LR";
  dagreGraph.setGraph({ rankdir: direction });

  nodes.forEach((node) => {
    const nodeWidth = nodeTypeSizes[node.type].width;
    const nodeHeight = nodeTypeSizes[node.type].height;
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const newNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    const nodeWidth = nodeTypeSizes[node.type].width;
    const nodeHeight = nodeTypeSizes[node.type].height;
    const newNode = {
      ...node,
      targetPosition: isHorizontal ? "left" : "top",
      sourcePosition: isHorizontal ? "right" : "bottom",
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };

    return newNode;
  });

  return { nodes: newNodes, edges };
};

function App() {
  const [prValue, setPrValue] = useState("");
  const handlePrValue = (e) => {
    setPrValue(e.target.value);
  };

  const [nodes, setNodes, onNodesChange] = useState([]);
  const [edges, setEdges, onEdgesChange] = useState([]);

  const handleSubmit = (event) => {
    event.preventDefault();
    const loadingToast = toast.loading("Tracking pull request...");
    fetch("/pr?id=" + prValue)
      .then((response) => response.json())
      .then((data) => {
        if (data["error"]) {
          toast.update(loadingToast, {
            render: data["error"],
            type: "error",
            isLoading: false,
            autoClose: 5000,
          });
          return;
        }
        console.log(data);
        toast.update(loadingToast, {
          render: "Pull request tracked!",
          type: "success",
          isLoading: false,
          autoClose: 1500,
        });

        const { nodes, edges } = flattenTree(data);
        const layoutedElements = getLayoutedElements(nodes, edges);
        setNodes(layoutedElements.nodes);
        setEdges(layoutedElements.edges);
      })
      .catch((error) => {
        toast.update(loadingToast, {
          render: "Failed to fetch data",
          type: "error",
          isLoading: false,
          autoClose: 5000,
        });
        console.error(error);
      });
  };

  const nodeTypes = { branch: BranchNode, pr: PRNode };

  return (
    <div style={{ width: "100vw", height: "100vh" }}>
      <ToastContainer />
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        fitView
        colorMode="light"
      >
        <Panel>
          <h2 className="nm">Nixpkgs PR Tracker</h2>
          <p className="nm">
            <small>
              <a href="https://github.com/humaidq/nixpkgs-pr-tracker">
                Source Code
              </a>
            </small>
          </p>
          <form onSubmit={handleSubmit}>
            <input
              id={"prId"}
              placeholder={"PR ID"}
              value={prValue}
              onChange={handlePrValue}
            ></input>
            <button className={"btn"}>Check</button>
          </form>
        </Panel>
        <Controls />
        <Background />
      </ReactFlow>
    </div>
  );
}

export default App;
