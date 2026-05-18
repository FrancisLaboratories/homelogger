import React, { useEffect, useState, useCallback } from "react";
import { Button, Form, ListGroup } from "react-bootstrap";
import "bootstrap-icons/font/bootstrap-icons.css";
import { SERVER_URL } from "@/context/DemoContext";
import type { Task } from "./TasksSection";
import TaskItem from "./TaskItem";
import AddTaskModal from "./AddTaskModal";
import applySort from "@/utils/taskSort";
import { Link } from "react-router-dom";

type SortOption =
  | "due_asc"
  | "due_desc"
  | "priority"
  | "created_desc"
  | "label_asc";
type FilterOption =
  | "active"
  | "completed"
  | "all"
  | "priority_high"
  | "recurring"
  | "no_date";

const PRIORITY_ORDER: Record<string, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
  "": 4,
};

const SPACE_URLS: Record<string, string> = {
  BuildingExterior: "/building-exterior",
  BuildingInterior: "/building-interior",
  Electrical: "/electrical",
  HVAC: "/hvac",
  Plumbing: "/plumbing",
  Yard: "/yard",
};

const SPACE_LABELS: Record<string, string> = {
  BuildingExterior: "Building Exterior",
  BuildingInterior: "Building Interior",
  Electrical: "Electrical",
  HVAC: "HVAC",
  Plumbing: "Plumbing",
  Yard: "Yard",
};

function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}

function isBeforeToday(dueDate: string): boolean {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return new Date(dueDate + "T00:00:00") < today;
}

function getCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(
    new RegExp("(?:^|; )" + name + "=([^;]*)"),
  );
  return match ? decodeURIComponent(match[1]) : null;
}

function setCookie(name: string, value: string) {
  const expires = new Date();
  expires.setFullYear(expires.getFullYear() + 1);
  document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires.toUTCString()}; path=/`;
}

interface TaskGroup {
  label: string;
  tasks: Task[];
  headerClass: string;
}

const TasksDashboard: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [applianceNames, setApplianceNames] = useState<Record<number, string>>(
    {},
  );
  const [quickLabel, setQuickLabel] = useState("");
  const [showAddModal, setShowAddModal] = useState(false);
  const [sortOption, setSortOption] = useState<SortOption>(
    () => (getCookie("hl_dashboard_sort") as SortOption) || "due_asc",
  );
  const [filterOption, setFilterOption] = useState<FilterOption>(
    () => (getCookie("hl_dashboard_filter") as FilterOption) || "active",
  );
  type GroupMode = "due" | "source" | "priority" | "none";
  const [groupMode, setGroupMode] = useState<GroupMode>(
    () => (getCookie("hl_dashboard_group_mode") as GroupMode) || "due",
  );

  const fetchTasks = useCallback(async () => {
    try {
      const res = await fetch(
        `${SERVER_URL}/task/dashboard?includeCompleted=true`,
      );
      if (!res.ok) return;
      const data: Task[] = await res.json();
      setTasks(data);

      const applianceIds = Array.from(
        new Set(
          data.filter((t) => t.applianceId).map((t) => Number(t.applianceId)),
        ),
      );
      const nameMap: Record<number, string> = {};
      await Promise.all(
        applianceIds.map(async (id) => {
          try {
            const r = await fetch(`${SERVER_URL}/appliances/${id}`);
            if (!r.ok) return;
            const a = await r.json();
            nameMap[id] = a.applianceName || `Appliance ${id}`;
          } catch (err) {
            console.warn(`Failed to fetch appliance ${id}:`, err);
          }
        }),
      );
      setApplianceNames(nameMap);
    } catch (e) {
      console.error("Error fetching dashboard tasks:", e);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleComplete = (updated: Task) => {
    setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)));
    fetchTasks();
    if (!updated.isRecurring && filterOption === "active") {
      setFilterOption("all");
      setCookie("hl_dashboard_filter", "all");
    }
  };

  const handleDelete = (id: number) => {
    setTasks((prev) => prev.filter((t) => t.id !== id));
  };

  const handleEdit = (updated: Task) => {
    setTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)));
  };

  const handleQuickAdd = async () => {
    const label = quickLabel.trim();
    if (!label) return;
    try {
      const res = await fetch(`${SERVER_URL}/task/add`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ label }),
      });
      if (!res.ok) throw new Error(await res.text());
      const created: Task = await res.json();
      setTasks((prev) => [...prev, created]);
      setQuickLabel("");
    } catch (e) {
      console.error("Error adding task:", e);
    }
  };

  function getSourceLabel(task: Task): string {
    if (task.applianceId != null) {
      return applianceNames[task.applianceId]
        ? `Appliance: ${applianceNames[task.applianceId]}`
        : `Appliance ${task.applianceId}`;
    }
    if (task.spaceType) return SPACE_LABELS[task.spaceType] || task.spaceType;
    return "General";
  }

  function getSourceHref(task: Task): string | undefined {
    if (task.applianceId != null) return `/appliance?id=${task.applianceId}`;
    if (task.spaceType) return SPACE_URLS[task.spaceType];
    return undefined;
  }

  function applyFilter(list: Task[]): Task[] {
    let result = list;
    if (filterOption === "active") result = result.filter((t) => !t.checked);
    else if (filterOption === "completed")
      result = result.filter((t) => t.checked);
    else if (filterOption === "priority_high")
      result = result.filter(
        (t) => t.priority === "high" || t.priority === "critical",
      );
    else if (filterOption === "recurring")
      result = result.filter((t) => t.isRecurring);
    else if (filterOption === "no_date")
      result = result.filter((t) => !t.dueDate);
    return result;
  }

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const in7 = addDays(today, 7);
  const in30 = addDays(today, 30);

  const filtered = applyFilter(tasks);

  const groups: TaskGroup[] = [
    {
      label: "Overdue",
      headerClass: "text-danger",
      tasks: applySort(
        filtered.filter((t) => t.dueDate && isBeforeToday(t.dueDate)),
        sortOption,
      ),
    },
    {
      label: "Due Next 7 Days",
      headerClass: "text-warning",
      tasks: applySort(
        filtered.filter((t) => {
          if (!t.dueDate) return false;
          const d = new Date(t.dueDate + "T00:00:00");
          return d >= today && d < in7;
        }),
        sortOption,
      ),
    },
    {
      label: "Due Next 30 Days",
      headerClass: "text-primary",
      tasks: applySort(
        filtered.filter((t) => {
          if (!t.dueDate) return false;
          const d = new Date(t.dueDate + "T00:00:00");
          return d >= in7 && d < in30;
        }),
        sortOption,
      ),
    },
    {
      label: "Later",
      headerClass: "text-muted",
      tasks: applySort(
        filtered.filter((t) => {
          if (!t.dueDate) return false;
          return new Date(t.dueDate + "T00:00:00") >= in30;
        }),
        sortOption,
      ),
    },
    {
      label: "No Due Date",
      headerClass: "text-muted",
      tasks: applySort(
        filtered.filter((t) => !t.dueDate),
        sortOption,
      ),
    },
  ];

  const totalActive = tasks.filter((t) => !t.checked).length;
  const totalFiltered = filtered.length;

  return (
    <div>
      <div className="d-flex align-items-center justify-content-between mb-3">
        <h4 className="mb-0">
          Tasks
          {totalActive > 0 && (
            <span
              className="text-muted ms-2"
              style={{ fontSize: "1rem", fontWeight: "normal" }}
            >
              ({totalActive} active)
            </span>
          )}
        </h4>
        <Button
          variant="outline-primary"
          size="sm"
          onClick={() => setShowAddModal(true)}
        >
          <i className="bi bi-plus-lg me-1" aria-hidden="true" /> Add Detailed
          Task
        </Button>
      </div>

      <div className="d-flex flex-wrap gap-2 mb-2 align-items-center">
        <Form.Select
          size="sm"
          style={{ width: "auto" }}
          value={filterOption}
          onChange={(e) => {
            setFilterOption(e.target.value as FilterOption);
            setCookie("hl_dashboard_filter", e.target.value);
          }}
          aria-label="Filter tasks"
        >
          <option value="active">Active</option>
          <option value="completed">Completed</option>
          <option value="all">All</option>
          <option value="priority_high">High / Critical priority</option>
          <option value="recurring">Recurring only</option>
          <option value="no_date">No due date</option>
        </Form.Select>

        <Form.Select
          size="sm"
          style={{ width: "auto" }}
          value={sortOption}
          onChange={(e) => {
            const v = e.target.value as SortOption;
            setSortOption(v);
            setCookie("hl_dashboard_sort", v);
          }}
          aria-label="Sort tasks"
        >
          <option value="due_asc">Due date ↑</option>
          <option value="due_desc">Due date ↓</option>
          <option value="priority">Priority</option>
          <option value="label_asc">Label A–Z</option>
          <option value="created_desc">Recently added</option>
        </Form.Select>

        <Form.Select
          size="sm"
          style={{ width: "auto" }}
          value={groupMode}
          onChange={(e) => {
            const v = e.target.value as GroupMode;
            setGroupMode(v);
            setCookie("hl_dashboard_group_mode", v);
          }}
          aria-label="Group tasks"
        >
          <option value="due">Group by due date</option>
          <option value="source">Group by source</option>
          <option value="priority">Group by priority</option>
          <option value="none">No grouping</option>
        </Form.Select>
      </div>

      <div className="d-flex gap-2 mb-3">
        <Form.Control
          type="text"
          size="sm"
          placeholder="Add a new task..."
          value={quickLabel}
          onChange={(e) => setQuickLabel(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleQuickAdd();
          }}
        />
        <Button variant="outline-secondary" size="sm" onClick={handleQuickAdd}>
          <i className="bi bi-plus-lg" aria-hidden="true" />
        </Button>
      </div>

      {totalFiltered === 0 && (
        <div className="text-muted">
          {filterOption === "active"
            ? "No active tasks. Great work!"
            : "No tasks match the current filter."}
        </div>
      )}
      {groupMode === "none"
        ? (() => {
            const flat = applySort(filtered, sortOption);
            if (flat.length === 0) return null;
            return (
              <ListGroup className="mb-2">
                {flat.map((task) => (
                  <TaskItem
                    key={task.id}
                    task={task}
                    onComplete={handleComplete}
                    onDelete={handleDelete}
                    onEdit={handleEdit}
                    showSource
                    sourceLabel={getSourceLabel(task)}
                    sourceHref={getSourceHref(task)}
                  />
                ))}
              </ListGroup>
            );
          })()
        : groupMode === "source"
          ? (() => {
              const sourceKeys = Array.from(
                new Set(filtered.map((t) => getSourceLabel(t))),
              ).sort((a, b) => a.localeCompare(b));
              return sourceKeys.map((src) => {
                const sourceTasks = applySort(
                  filtered.filter((t) => getSourceLabel(t) === src),
                  sortOption,
                );
                if (sourceTasks.length === 0) return null;
                const href = getSourceHref(sourceTasks[0]);
                return (
                  <div key={src} className="mb-3">
                    <h6 className="mb-1" style={{ fontWeight: 600 }}>
                      {href ? (
                        <Link to={href} className="text-decoration-none">
                          {src}
                        </Link>
                      ) : (
                        src
                      )}
                      <span
                        className="ms-2 text-muted fw-normal"
                        style={{ fontSize: "0.85rem" }}
                      >
                        ({sourceTasks.length})
                      </span>
                    </h6>
                    <ListGroup>
                      {sourceTasks.map((task) => (
                        <TaskItem
                          key={task.id}
                          task={task}
                          onComplete={handleComplete}
                          onDelete={handleDelete}
                          onEdit={handleEdit}
                        />
                      ))}
                    </ListGroup>
                  </div>
                );
              });
            })()
          : groupMode === "priority"
            ? (() => {
                const prioKeys = Object.keys(PRIORITY_ORDER).sort(
                  (a, b) => PRIORITY_ORDER[a] - PRIORITY_ORDER[b],
                );
                return prioKeys.map((key) => {
                  const label = key
                    ? key.charAt(0).toUpperCase() + key.slice(1)
                    : "No priority";
                  const prioTasks = applySort(
                    filtered.filter((t) => (t.priority || "") === key),
                    sortOption,
                  );
                  if (prioTasks.length === 0) return null;
                  return (
                    <div key={key || "none"} className="mb-3">
                      <h6 className="mb-1" style={{ fontWeight: 600 }}>
                        {label}
                        <span
                          className="ms-2 text-muted fw-normal"
                          style={{ fontSize: "0.85rem" }}
                        >
                          ({prioTasks.length})
                        </span>
                      </h6>
                      <ListGroup>
                        {prioTasks.map((task) => (
                          <TaskItem
                            key={task.id}
                            task={task}
                            onComplete={handleComplete}
                            onDelete={handleDelete}
                            onEdit={handleEdit}
                          />
                        ))}
                      </ListGroup>
                    </div>
                  );
                });
              })()
            : groups.map((group) => {
                if (group.tasks.length === 0) return null;
                return (
                  <div key={group.label} className="mb-3">
                    <h6
                      className={`${group.headerClass} mb-1`}
                      style={{ fontWeight: 600 }}
                    >
                      {group.label}
                      <span
                        className="ms-2 text-muted fw-normal"
                        style={{ fontSize: "0.85rem" }}
                      >
                        ({group.tasks.length})
                      </span>
                    </h6>
                    <ListGroup>
                      {group.tasks.map((task) => (
                        <TaskItem
                          key={task.id}
                          task={task}
                          onComplete={handleComplete}
                          onDelete={handleDelete}
                          onEdit={handleEdit}
                          showSource
                          sourceLabel={getSourceLabel(task)}
                          sourceHref={getSourceHref(task)}
                        />
                      ))}
                    </ListGroup>
                  </div>
                );
              })}

      <AddTaskModal
        show={showAddModal}
        onHide={() => setShowAddModal(false)}
        onAdd={(task) => {
          setTasks((prev) => [...prev, task]);
          setShowAddModal(false);
        }}
        startDetailed={true}
      />
    </div>
  );
};

export default TasksDashboard;
