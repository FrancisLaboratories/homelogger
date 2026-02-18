"use client";

import React, { useEffect, useState } from 'react';
import { Badge, Card, Col, Container, Form, Row } from 'react-bootstrap';
import "bootstrap-icons/font/bootstrap-icons.css";
import MyNavbar from '../components/Navbar';
import ListGroup from 'react-bootstrap/ListGroup';
import TodoItem from '../components/TodoItem';
import { SERVER_URL } from "@/pages/_app";

const todoUrl = `${SERVER_URL}/todo`;
const todoAddUrl = `${SERVER_URL}/todo/add`;

type Todo = {
  id: string | number;
  label: string;
  checked: boolean;
  applianceId?: number | null;
  spaceType?: string | null;
  sourceLabel?: string | null;
  createdAt?: string | null;
  CreatedAt?: string | null;
  created_at?: string | null;
};

type BudgetScenario = {
  id: number;
  name: string;
  startDate: string;
  horizonMonths: number;
  inflationRate: number;
  isActive: boolean;
  notes: string;
};

type DashboardSummary = {
  scenario?: BudgetScenario | null;
  horizonMonths: number;
  monthlySavings: number;
  plannedCostTotal: number;
  plannedCostCount: number;
  upcoming30DaysTotal: number;
  overdueTotal: number;
  overdueCount: number;
  upgradeCount: number;
  upgradeTotal: number;
  recurringCount: number;
  recurringDue30: number;
  repairTotal: number;
  maintenanceTotal: number;
};

const HomePage: React.FC = () => {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [sortOption, setSortOption] = useState<string>('created_desc');
  const [filterOption, setFilterOption] = useState<string>('not_completed');
  const [groupBySource, setGroupBySource] = useState<boolean>(false);
  const [scenarios, setScenarios] = useState<BudgetScenario[]>([]);
  const [selectedScenarioId, setSelectedScenarioId] = useState<number | null>(null);
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const highUpcomingMultiplier = 1.5;

  const prettySpace = (s?: string | null) => {
    if (!s) return null;
    switch (s) {
      case 'BuildingExterior': return 'Building Exterior';
      case 'BuildingInterior': return 'Building Interior';
      case 'Electrical': return 'Electrical';
      case 'HVAC': return 'HVAC';
      case 'Plumbing': return 'Plumbing';
      case 'Yard': return 'Yard';
      default:
        return s.replace(/([a-z])([A-Z])/g, '$1 $2');
    }
  };

  useEffect(() => {
    try {
      const savedSort = localStorage.getItem('homelogger_todo_sort');
      const savedFilter = localStorage.getItem('homelogger_todo_filter');
      const savedGroup = localStorage.getItem('homelogger_todo_group');
      if (savedSort) setSortOption(savedSort);
      if (savedFilter) setFilterOption(savedFilter);
      if (savedGroup) setGroupBySource(savedGroup === 'true');
    } catch (e) {
      // ignore
    }
  }, []);

  useEffect(() => {
    const fetchTodos = async () => {
      try {
        const response = await fetch(todoUrl);
        const data = await response.json();

        const dataTyped: Todo[] = data as Todo[];
        const applianceIds: number[] = Array.from(new Set(dataTyped.filter((t) => t.applianceId).map((t) => Number(t.applianceId))));
        const nameMap: Record<number, string> = {};
        await Promise.all(applianceIds.map(async (id) => {
          try {
            const r = await fetch(`${SERVER_URL}/appliances/${id}`);
            if (!r.ok) return;
            const a = await r.json();
            nameMap[id] = a.applianceName || `Appliance ${id}`;
          } catch (e) {
            console.error('Error loading appliance name', e);
          }
        }));

        const enriched: Todo[] = dataTyped.map((t) => ({
          ...t,
          sourceLabel: t.applianceId ? nameMap[Number(t.applianceId)] : prettySpace(t.spaceType || null),
        }));

        setTodos(enriched);
      } catch (error) {
        console.error('Error fetching todos:', error);
      }
    };

    fetchTodos();
  }, []);

  useEffect(() => {
    const loadScenarios = async () => {
      try {
        const response = await fetch(`${SERVER_URL}/budget/scenarios`);
        if (!response.ok) return;
        const data = await response.json();
        setScenarios(data);
        if (data.length > 0 && selectedScenarioId === null) {
          const active = data.find((s: BudgetScenario) => s.isActive) || data[0];
          setSelectedScenarioId(active.id);
        }
      } catch (error) {
        console.error('Error fetching scenarios:', error);
      }
    };

    loadScenarios();
  }, []);

  useEffect(() => {
    const loadSummary = async () => {
      try {
        let url = `${SERVER_URL}/dashboard/summary`;
        if (selectedScenarioId) {
          url = `${url}?scenarioId=${selectedScenarioId}`;
        }
        const response = await fetch(url);
        if (!response.ok) return;
        const data = await response.json();
        setSummary(data);
      } catch (error) {
        console.error('Error fetching dashboard summary:', error);
      }
    };

    loadSummary();
  }, [selectedScenarioId]);

  const highUpcoming = summary && summary.monthlySavings > 0
    ? summary.upcoming30DaysTotal > summary.monthlySavings * highUpcomingMultiplier
    : false;

  const handleAddTodo = async () => {
    const label = prompt('What should go in this item?');
    if (label) {
      const newTodo = { label, checked: false, userid: "1" };

      try {
        const response = await fetch(todoAddUrl, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(newTodo),
        });

        if (!response.ok) {
          throw new Error('Failed to add todo');
        }

        const addedTodo = await response.json();
        setTodos([...todos, addedTodo]);
      } catch (error) {
        console.error('Error adding todo:', error);
      }
    }
  };

  const handleDeleteTodo = (id: string) => {
    setTodos((prevTodos) => prevTodos.filter((todo) => String(todo.id) !== id));
  };

  const handleToggleTodo = (id: string, checked: boolean) => {
    setTodos((prevTodos) => prevTodos.map(t => (String(t.id) === String(id) ? { ...t, checked } : t)));
  };

  return (
    <Container>
      <MyNavbar />
      <Row className="g-3" style={{ marginTop: '8px' }}>
        <Col lg={3} md={6}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Budget Snapshot</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.monthlySavings || 0).toFixed(2)}/mo</div>
              <div style={{ fontSize: '0.85rem' }}>
                {summary?.scenario ? `${summary.scenario.name} • ${summary.horizonMonths || summary.scenario.horizonMonths} mo` : 'No scenario'}
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col lg={3} md={6}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Upcoming (30 days)</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>${(summary?.upcoming30DaysTotal || 0).toFixed(2)}</div>
              <div style={{ fontSize: '0.85rem' }}>
                {summary?.plannedCostCount || 0} planned items
                {(summary?.upcoming30DaysTotal || 0) > 0 && (
                  <Badge bg="warning" text="dark" style={{ marginLeft: 8 }}>Upcoming</Badge>
                )}
                {(summary?.overdueCount || 0) > 0 && (
                  <Badge bg="danger" style={{ marginLeft: 8 }}>Overdue</Badge>
                )}
                {highUpcoming && (
                  <Badge bg="danger" style={{ marginLeft: 8 }}>High 30-day</Badge>
                )}
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col lg={3} md={6}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Upgrades</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>{summary?.upgradeCount ?? upgrades.length} projects</div>
              <div style={{ fontSize: '0.85rem' }}>${(summary?.upgradeTotal ?? 0).toFixed(2)} planned</div>
            </Card.Body>
          </Card>
        </Col>
        <Col lg={3} md={6}>
          <Card>
            <Card.Body>
              <div style={{ fontSize: '0.9rem', color: '#6c757d' }}>Recurring Due Soon</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 600 }}>{summary?.recurringDue30 ?? 0} tasks</div>
              <div style={{ fontSize: '0.85rem' }}>
                {summary?.recurringCount ?? 0} total
                {(summary?.recurringDue30 || 0) > 0 && (
                  <Badge bg="danger" style={{ marginLeft: 8 }}>Due Soon</Badge>
                )}
              </div>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 12 }}>
        <h4 id='maintext' style={{ marginBottom: 0 }}>To-dos</h4>
        <Form.Select
          aria-label="Budget scenario"
          value={selectedScenarioId ?? ''}
          onChange={(e) => setSelectedScenarioId(e.target.value ? Number(e.target.value) : null)}
          style={{ maxWidth: '260px' }}
        >
          <option value="">Budget scenario</option>
          {scenarios.map((s) => (
            <option key={s.id} value={s.id}>{s.name}</option>
          ))}
        </Form.Select>
      </div>

      <div style={{display: 'flex', gap: '8px', marginBottom: '8px'}}>
        <Form.Select aria-label="Sort todos" value={sortOption} onChange={(e) => { setSortOption(e.target.value); try { localStorage.setItem('homelogger_todo_sort', e.target.value); } catch {} }} style={{maxWidth: '220px'}}>
          <option value="created_desc">Created (newest)</option>
          <option value="created_asc">Created (oldest)</option>
          <option value="label_asc">Label (A - Z)</option>
          <option value="label_desc">Label (Z - A)</option>
        </Form.Select>

        <Form.Select aria-label="Filter todos" value={filterOption} onChange={(e) => { setFilterOption(e.target.value); try { localStorage.setItem('homelogger_todo_filter', e.target.value); } catch {} }} style={{maxWidth: '180px'}}>
          <option value="all">All</option>
          <option value="completed">Completed</option>
          <option value="not_completed">Not completed</option>
        </Form.Select>

        <Form.Check type="checkbox" label="Group by appliance / space" checked={groupBySource} onChange={(e) => { setGroupBySource(e.target.checked); try { localStorage.setItem('homelogger_todo_group', e.target.checked ? 'true' : 'false'); } catch {} }} style={{alignSelf: 'center', marginLeft: '8px'}} />
      </div>

      <ListGroup>
        {(() => {
          const filtered = todos.filter((t) => {
            if (filterOption === 'all') return true;
            if (filterOption === 'completed') return !!t.checked;
            if (filterOption === 'not_completed') return !t.checked;
            return true;
          });

          const comparator = (a: Todo, b: Todo) => {
            const sa = (a.label || '').toString();
            const sb = (b.label || '').toString();

            const ca = a.createdAt || a.CreatedAt || a.created_at || null;
            const cb = b.createdAt || b.CreatedAt || b.created_at || null;

            if (sortOption === 'label_asc') return sa.localeCompare(sb);
            if (sortOption === 'label_desc') return sb.localeCompare(sa);

            const da = ca ? new Date(ca).getTime() : 0;
            const db = cb ? new Date(cb).getTime() : 0;
            if (sortOption === 'created_asc') return da - db || sa.localeCompare(sb);
            return db - da || sa.localeCompare(sb);
          };

          if (groupBySource) {
            const groups: Record<string, Todo[]> = {};
            filtered.forEach((t) => {
              const key = t.sourceLabel || prettySpace(t.spaceType || null) || 'General';
              if (!groups[key]) groups[key] = [];
              groups[key].push(t);
            });

            const keys = Object.keys(groups).sort();
            return keys.map((k) => (
              <React.Fragment key={k}>
                <ListGroup.Item className="fw-bold">{k}</ListGroup.Item>
                {groups[k].slice().sort(comparator).map((todo) => (
                  <TodoItem key={String(todo.id)} id={String(todo.id)} label={todo.label} checked={todo.checked} onDelete={handleDeleteTodo} onToggle={handleToggleTodo} applianceId={todo.applianceId || undefined} spaceType={todo.spaceType || undefined} sourceLabel={todo.sourceLabel || undefined} createdAt={todo.createdAt || todo.CreatedAt || todo.created_at || undefined} />
                ))}
              </React.Fragment>
            ));
          }

          const sorted = filtered.slice().sort(comparator);
          return sorted.map((todo, index) => (
            <TodoItem key={index} id={String(todo.id)} label={todo.label} checked={todo.checked} onDelete={handleDeleteTodo} onToggle={handleToggleTodo} applianceId={todo.applianceId || undefined} spaceType={todo.spaceType || undefined} sourceLabel={todo.sourceLabel || undefined} createdAt={todo.createdAt || todo.CreatedAt || todo.created_at || undefined} />
          ));
        })()}
      </ListGroup>
      <i className="bi bi-plus-square-fill" onClick={handleAddTodo} style={{ fontSize: '2rem', cursor: "pointer" }}></i>
    </Container>
  );
};

export default HomePage;
