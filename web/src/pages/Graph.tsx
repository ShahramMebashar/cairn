import { useMemo } from "react";
import {
  Background,
  Controls,
  Handle,
  Position,
  ReactFlow,
  type Edge,
  type Node,
  type NodeProps,
} from "@xyflow/react";
import dagre from "@dagrejs/dagre";
import "@xyflow/react/dist/style.css";
import { ArrowLeft, Network } from "lucide-react";
import { Button } from "@/components/ui/button";
import { StatusIcon } from "@/components/StatusIcon";
import { EmptyState } from "@/components/EmptyState";
import { useTasks } from "@/lib/queries";
import { cn } from "@/lib/utils";
import type { Status, Task } from "@/lib/api";

type NodeData = { task: Task; status: Status; ready: boolean; closedState: boolean };

const NODE_W = 230;
const NODE_H = 46;

function layout(tasks: Task[], status: Status): { nodes: Node<NodeData>[]; edges: Edge[] } {
  const ids = new Set(tasks.map((t) => t.id));
  const closed = new Set(status.closed ?? []);

  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "LR", nodesep: 22, ranksep: 70, marginx: 16, marginy: 16 });
  tasks.forEach((t) => g.setNode(t.id, { width: NODE_W, height: NODE_H }));

  const edges: Edge[] = [];
  for (const t of tasks) {
    for (const dep of t.deps ?? []) {
      if (!ids.has(dep)) continue; // dangling guarded elsewhere
      g.setEdge(dep, t.id);
      edges.push({
        id: `${dep}->${t.id}`,
        source: dep,
        target: t.id,
        animated: t.ready && !closed.has(t.status),
      });
    }
  }
  dagre.layout(g);

  const nodes: Node<NodeData>[] = tasks.map((t) => {
    const p = g.node(t.id);
    return {
      id: t.id,
      type: "task",
      position: { x: p.x - NODE_W / 2, y: p.y - NODE_H / 2 },
      data: { task: t, status, ready: t.ready, closedState: closed.has(t.status) },
    };
  });
  return { nodes, edges };
}

function TaskNode({ data }: NodeProps<Node<NodeData>>) {
  const { task, status, ready, closedState } = data;
  return (
    <div
      className={cn(
        "flex h-[46px] w-[230px] items-center gap-2 rounded-lg border bg-panel px-2.5 shadow-xs",
        closedState && "opacity-70",
        ready && !closedState && "border-brand/60 ring-1 ring-brand/30",
      )}
    >
      <Handle type="target" position={Position.Left} className="!size-1.5 !bg-border" />
      <StatusIcon status={task.status} closed={status.closed} initial={status.initial} className="size-3.5" />
      <div className="min-w-0">
        <div className="font-mono text-[10px] leading-tight text-muted-foreground">{task.id}</div>
        <div className="truncate text-xs leading-tight">{task.title}</div>
      </div>
      <Handle type="source" position={Position.Right} className="!size-1.5 !bg-border" />
    </div>
  );
}

const nodeTypes = { task: TaskNode };

export function Graph({
  path,
  status,
  onOpenTask,
  onBack,
}: {
  path: string;
  status: Status;
  onOpenTask: (id: string) => void;
  onBack: () => void;
}) {
  const { data: tasks } = useTasks(path);
  const { nodes, edges } = useMemo(() => layout(tasks ?? [], status), [tasks, status]);

  return (
    <div className="flex h-full flex-col">
      <header className="flex h-11 shrink-0 items-center gap-2 border-b px-3">
        <Button variant="ghost" size="icon" aria-label="Back" onClick={onBack}>
          <ArrowLeft />
        </Button>
        <span className="text-sm font-medium">Dependency graph</span>
        <span className="text-xs text-muted-foreground">{nodes.length} tasks</span>
      </header>
      <div className="min-h-0 flex-1">
        {nodes.length === 0 ? (
          <EmptyState
            icon={Network}
            title="No tasks to graph yet"
            message="As tasks and their dependencies appear, you'll see the dependency graph here."
          />
        ) : (
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            fitView
            proOptions={{ hideAttribution: true }}
            onNodeClick={(_, n) => onOpenTask(n.id)}
            nodesDraggable={false}
          >
            <Background gap={16} className="!bg-app" />
            <Controls showInteractive={false} />
          </ReactFlow>
        )}
      </div>
    </div>
  );
}
