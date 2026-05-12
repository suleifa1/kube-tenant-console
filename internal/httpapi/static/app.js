let state = {};
let clusterState = null;
let roleDraftRules = [];

const RBAC_VERBS = [
  { value: "get", zone: "read" },
  { value: "list", zone: "read" },
  { value: "watch", zone: "read" },
  { value: "create", zone: "write" },
  { value: "update", zone: "write" },
  { value: "patch", zone: "write" },
  { value: "delete", zone: "write" },
  { value: "deletecollection", zone: "write" },
  { value: "use", zone: "privileged" },
  { value: "bind", zone: "privileged" },
  { value: "escalate", zone: "privileged" },
  { value: "approve", zone: "privileged" },
  { value: "sign", zone: "privileged" },
  { value: "impersonate", zone: "blocked", blocked: true },
];

const RBAC_CATALOG = [
  { group: "", label: "core", zone: "workload", resources: [
    "pods", "pods/log", "pods/status", "pods/exec!", "pods/attach!", "pods/portforward!",
    "services", "services/proxy", "endpoints", "configmaps", "secrets!", "serviceaccounts",
    "persistentvolumeclaims", "persistentvolumes!", "events", "namespaces",
    "resourcequotas", "limitranges", "replicationcontrollers",
  ] },
  { group: "apps", label: "apps", zone: "workload", resources: [
    "deployments", "deployments/scale", "deployments/status", "replicasets", "replicasets/scale",
    "replicasets/status", "statefulsets", "statefulsets/scale", "statefulsets/status",
    "daemonsets", "daemonsets/status", "controllerrevisions",
  ] },
  { group: "batch", label: "batch", zone: "workload", resources: [
    "jobs", "jobs/status", "cronjobs", "cronjobs/status",
  ] },
  { group: "networking.k8s.io", label: "networking", zone: "network", resources: [
    "ingresses", "ingresses/status", "networkpolicies", "ingressclasses",
  ] },
  { group: "autoscaling", label: "autoscaling", zone: "workload", resources: [
    "horizontalpodautoscalers", "horizontalpodautoscalers/status",
  ] },
  { group: "policy", label: "policy", zone: "policy", resources: [
    "poddisruptionbudgets", "poddisruptionbudgets/status", "podsecuritypolicies",
  ] },
  { group: "rbac.authorization.k8s.io", label: "rbac", zone: "blocked", resources: [
    "roles!", "rolebindings!", "clusterroles!", "clusterrolebindings!",
  ] },
  { group: "storage.k8s.io", label: "storage", zone: "platform", resources: [
    "storageclasses", "csidrivers", "csinodes", "csistoragecapacities", "volumeattachments",
  ] },
  { group: "coordination.k8s.io", label: "coordination", zone: "platform", resources: [
    "leases",
  ] },
  { group: "discovery.k8s.io", label: "discovery", zone: "network", resources: [
    "endpointslices",
  ] },
  { group: "admissionregistration.k8s.io", label: "admission", zone: "platform", resources: [
    "validatingwebhookconfigurations", "mutatingwebhookconfigurations", "validatingadmissionpolicies",
    "validatingadmissionpolicybindings",
  ] },
  { group: "apiextensions.k8s.io", label: "apiextensions", zone: "platform", resources: [
    "customresourcedefinitions", "customresourcedefinitions/status",
  ] },
  { group: "certificates.k8s.io", label: "certificates", zone: "blocked", resources: [
    "certificatesigningrequests!",
  ] },
  { group: "events.k8s.io", label: "events", zone: "read", resources: [
    "events",
  ] },
  { group: "node.k8s.io", label: "node", zone: "blocked", resources: [
    "runtimeclasses", "nodes!",
  ] },
  { group: "scheduling.k8s.io", label: "scheduling", zone: "platform", resources: [
    "priorityclasses",
  ] },
  { group: "__custom", label: "custom", zone: "custom", resources: [] },
];

const ROLE_PRESETS = [
  {
    id: "deployer",
    name: "Deployer",
    zone: "write",
    description: "Deployments and app config, read pods and rollout state.",
    rules: [
      rule("apps", ["deployments", "deployments/scale"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
      rule("apps", ["replicasets", "daemonsets", "statefulsets"], ["get", "list", "watch"]),
      rule("", ["pods", "pods/log", "services", "configmaps"], ["get", "list", "watch"]),
      rule("", ["services", "configmaps"], ["create", "update", "patch", "delete"]),
      rule("batch", ["jobs", "cronjobs"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
    ],
  },
  {
    id: "viewer",
    name: "Viewer",
    zone: "read",
    description: "Read common workload and network objects without secrets.",
    rules: [
      rule("", ["pods", "pods/log", "services", "endpoints", "configmaps", "resourcequotas", "limitranges", "events"], ["get", "list", "watch"]),
      rule("apps", ["deployments", "replicasets", "daemonsets", "statefulsets"], ["get", "list", "watch"]),
      rule("batch", ["jobs", "cronjobs"], ["get", "list", "watch"]),
      rule("networking.k8s.io", ["ingresses", "networkpolicies"], ["get", "list", "watch"]),
    ],
  },
  {
    id: "operator",
    name: "Namespace Operator",
    zone: "write",
    description: "Manage workloads, services, config, jobs, ingress, network policy, and HPA.",
    rules: [
      rule("", ["pods", "pods/log", "services", "endpoints", "configmaps", "persistentvolumeclaims", "resourcequotas", "limitranges"], ["get", "list", "watch"]),
      rule("", ["services", "configmaps", "persistentvolumeclaims"], ["create", "update", "patch", "delete"]),
      rule("apps", ["deployments", "deployments/scale", "replicasets", "statefulsets", "statefulsets/scale", "daemonsets"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
      rule("batch", ["jobs", "cronjobs"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
      rule("networking.k8s.io", ["ingresses", "networkpolicies"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
      rule("autoscaling", ["horizontalpodautoscalers"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
    ],
  },
  {
    id: "ci",
    name: "CI Bot",
    zone: "write",
    description: "Patch and roll out workloads from automation without interactive pod access.",
    rules: [
      rule("apps", ["deployments", "deployments/scale", "statefulsets", "statefulsets/scale"], ["get", "list", "watch", "create", "update", "patch"]),
      rule("batch", ["jobs"], ["get", "list", "watch", "create", "delete"]),
      rule("", ["pods", "pods/log", "services", "configmaps"], ["get", "list", "watch"]),
      rule("", ["configmaps"], ["create", "update", "patch"]),
    ],
  },
  {
    id: "networker",
    name: "Network Maintainer",
    zone: "network",
    description: "Manage ingress and namespace network policies.",
    rules: [
      rule("networking.k8s.io", ["ingresses", "networkpolicies"], ["get", "list", "watch", "create", "update", "patch", "delete"]),
      rule("", ["services", "endpoints"], ["get", "list", "watch"]),
    ],
  },
  {
    id: "custom",
    name: "Custom",
    zone: "custom",
    description: "Use the advanced rule builder below.",
    rules: [],
  },
];

const $ = (selector, root = document) => root.querySelector(selector);
const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

function rule(apiGroup, resources, verbs) {
  return { apiGroups: [apiGroup], resources, verbs };
}

async function api(path, body, method = "") {
  const requestMethod = method || (body ? "POST" : "GET");
  const res = await fetch(path, {
    method: requestMethod,
    headers: body ? { "content-type": "application/json" } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || "Request failed");
  return data;
}

async function refresh() {
  state = normalizeState(await api("/api/state"));
  render();
}

function render() {
  renderOptions();
  renderNamespaceNamePreview();
  renderAccessOptions();
  renderList("#tenantList", state.tenants, (tenant) => `
    <div class="item">
      <strong>${escapeHTML(tenant.name)}</strong>
      <span class="meta">prefix ${escapeHTML(tenant.namespacePrefix || tenant.name)}</span>
      <div class="item-actions">
        <button type="button" class="ghost danger-action state-delete-object" data-kind="tenant" data-path="/api/tenants/${escapeAttr(encodeURIComponent(tenant.id))}">Delete local tree</button>
      </div>
    </div>
  `);
  renderList("#namespaceList", state.namespaces, (ns) => `
    <div class="item">
      <strong>${escapeHTML(ns.name)}</strong>
      <span class="meta">quota ${escapeHTML(ns.quota.requestsCpu)} CPU / ${escapeHTML(ns.quota.requestsMemory)} memory / ${escapeHTML(ns.quota.pods)} pods</span>
      <div class="item-actions">
        <button class="ghost kube-ensure-namespace" data-namespace-id="${escapeAttr(ns.id)}">Create in cluster</button>
        <button class="ghost preview-button" data-kind="namespace" data-namespace-id="${escapeAttr(ns.id)}">Preview YAML</button>
        <button type="button" class="ghost danger-action state-delete-object" data-kind="namespace" data-path="/api/namespaces/${escapeAttr(encodeURIComponent(ns.id))}">Delete local</button>
      </div>
    </div>
  `);
  renderList("#roleList", state.roles, (role) => `
    <div class="item">
      <strong>${escapeHTML(role.name)}</strong>
      <span class="meta">${escapeHTML(role.scope)} scope, ${role.rules.length} rule(s)</span>
      ${role.scope === "cluster" ? `<span class="meta danger">Cluster role preview is not exposed by the current backend.</span>` : roleNamespacePicker(role)}
      <div class="item-actions">
        <button type="button" class="ghost danger-action state-delete-object" data-kind="role" data-path="/api/roles/${escapeAttr(encodeURIComponent(role.id))}">Delete local</button>
      </div>
    </div>
  `);
  renderAssignmentList();
  renderKubeconfigList();
  renderManifests();
  bindKubeActionButtons();
  bindKubeconfigIssueButtons();
  bindLocalDeleteButtons();
}

function renderOptions() {
  $$("select[name=tenantId]").forEach((select) => {
    select.innerHTML = optionList(state.tenants, "Select tenant", (tenant) => tenant.name);
  });
  $$("select[name=namespaceId]").forEach((select) => {
    if (select.id === "assignmentNamespace" || select.id === "kubeconfigNamespace" || select.id === "serviceAccountNamespace") return;
    select.innerHTML = optionList(state.namespaces, "Select namespace", (ns) => ns.name);
  });
  $$("select[name=roleId]").forEach((select) => {
    if (select.id === "assignmentRole") return;
    select.innerHTML = optionList(state.roles, "Select role", (role) => role.name);
  });
}

function renderNamespaceNamePreview() {
  const tenantSelect = $("#namespaceTenant");
  const suffixInput = $("#namespaceSuffix");
  const prefixNode = $("#namespacePrefixPreview");
  const fullNode = $("#namespaceFullNamePreview");
  if (!tenantSelect || !suffixInput || !prefixNode || !fullNode) return;

  const tenant = state.tenants.find((item) => item.id === tenantSelect.value);
  const suffix = suffixInput.value.trim();
  if (!tenant) {
    prefixNode.textContent = "select tenant";
    fullNode.textContent = "select tenant";
    return;
  }

  const prefix = namespacePrefix(tenant);
  prefixNode.textContent = `${prefix}-`;
  fullNode.textContent = suffix ? joinNamespaceName(prefix, suffix) : `${prefix}-...`;
}

function renderAccessOptions() {
  const namespaceSelect = $("#assignmentNamespace");
  const roleSelect = $("#assignmentRole");
  if (!namespaceSelect || !roleSelect) return;

  const previousNamespace = namespaceSelect.value;
  namespaceSelect.innerHTML = optionList(state.namespaces, "Select namespace", (ns) => ns.name);
  if (state.namespaces.some((ns) => ns.id === previousNamespace)) {
    namespaceSelect.value = previousNamespace;
  }

  renderAssignmentRoles();
  renderSubjectHint();
  renderServiceAccountOptions();
  renderKubeconfigOptions();
}

function renderAssignmentRoles() {
  const namespaceSelect = $("#assignmentNamespace");
  const roleSelect = $("#assignmentRole");
  if (!namespaceSelect || !roleSelect) return;

  const namespace = state.namespaces.find((ns) => ns.id === namespaceSelect.value);
  const roles = namespace
    ? state.roles.filter((role) => role.tenantId === namespace.tenantId && role.scope !== "cluster")
    : [];

  roleSelect.innerHTML = optionList(roles, namespace ? "Select role" : "Select namespace first", (role) => role.name);
  roleSelect.disabled = !namespace || roles.length === 0;
}

function renderSubjectHint() {
  const kind = $("#subjectKind")?.value || "User";
  const hint = $("#subjectKindHint");
  const inputLabel = $("#subjectNameInputLabel");
  const input = $("#subjectNameInput");
  const selectLabel = $("#subjectNameSelectLabel");
  const select = $("#subjectNameSelect");
  if (!hint || !inputLabel || !input || !selectLabel || !select) return;

  const details = {
    User: {
      text: "External authenticated user from GKE/OIDC/IAM. Kubernetes does not create this user.",
      placeholder: "alice@example.com",
    },
    Group: {
      text: "External group from the identity provider. Use it for team access.",
      placeholder: "developers@example.com",
    },
    ServiceAccount: {
      text: "Managed ServiceAccount in the selected namespace. Use it for bots and kubeconfigs.",
      placeholder: "ci-deploy",
    },
  }[kind];

  hint.textContent = details.text;
  input.placeholder = details.placeholder;
  if (kind === "ServiceAccount") {
    inputLabel.hidden = true;
    input.disabled = true;
    input.required = false;
    selectLabel.hidden = false;
    select.disabled = false;
    select.required = true;
    renderSubjectServiceAccountSelect();
  } else {
    inputLabel.hidden = false;
    input.disabled = false;
    input.required = true;
    selectLabel.hidden = true;
    select.disabled = true;
    select.required = false;
    select.innerHTML = "";
  }
}

function renderSubjectServiceAccountSelect() {
  const select = $("#subjectNameSelect");
  if (!select) return;

  const namespaceId = $("#assignmentNamespace")?.value || "";
  const accounts = serviceAccountsForNamespace(namespaceId);
  if (!namespaceId) {
    select.innerHTML = `<option value="">Select namespace first</option>`;
    select.disabled = true;
    return;
  }
  if (!accounts.length) {
    select.innerHTML = `<option value="">Create service account first</option>`;
    select.disabled = true;
    return;
  }

  select.disabled = false;
  select.innerHTML = [
    `<option value="">Select service account</option>`,
    ...accounts.map((account) => `<option value="${escapeAttr(account.name)}">${escapeHTML(account.name)}</option>`),
  ].join("");
}

function renderServiceAccountOptions() {
  const namespaceSelect = $("#serviceAccountNamespace");
  if (!namespaceSelect) return;

  const previousNamespace = namespaceSelect.value;
  namespaceSelect.innerHTML = optionList(state.namespaces, "Select namespace", (ns) => ns.name);
  if (state.namespaces.some((ns) => ns.id === previousNamespace)) {
    namespaceSelect.value = previousNamespace;
  }

  renderServiceAccountList();
}

function renderServiceAccountList() {
  const target = $("#serviceAccountList");
  if (!target) return;

  const namespaceId = $("#serviceAccountNamespace")?.value || "";
  if (!namespaceId) {
    target.innerHTML = `<p class="meta">Select namespace to see service accounts.</p>`;
    return;
  }

  const namespace = state.namespaces.find((ns) => ns.id === namespaceId);
  const accounts = serviceAccountsForNamespace(namespaceId);
  target.innerHTML = accounts.length
    ? accounts.map((account) => `
      <div class="item compact-item">
        <strong>${escapeHTML(account.name)}</strong>
        <span class="meta">${escapeHTML(namespace?.name || "unknown namespace")}</span>
        <div class="item-actions">
          <button type="button" class="ghost kube-ensure-service-account" data-namespace-id="${escapeAttr(namespaceId)}" data-sa-id="${escapeAttr(account.id)}">Create in cluster</button>
          <button type="button" class="ghost preview-button" data-kind="serviceaccount" data-namespace-id="${escapeAttr(namespaceId)}" data-sa-id="${escapeAttr(account.id)}">Preview YAML</button>
          <button type="button" class="ghost danger-action state-delete-object" data-kind="serviceaccount" data-path="/api/serviceaccounts/${escapeAttr(encodeURIComponent(account.id))}">Delete local</button>
        </div>
      </div>
    `).join("")
    : `<p class="meta">No service accounts in this namespace yet.</p>`;
  bindPreviewButtons();
  bindKubeActionButtons();
  bindLocalDeleteButtons();
}

function renderAssignmentList() {
  renderList("#assignmentList", state.assignments, (assignment) => {
    const namespace = state.namespaces.find((ns) => ns.id === assignment.namespaceId);
    const role = state.roles.find((item) => item.id === assignment.roleId);
    return `
      <div class="item">
        <strong>${escapeHTML(role?.name || "missing role")}</strong>
        <span class="meta">${escapeHTML(assignment.subjectKind)} ${escapeHTML(assignment.subjectName)}${namespace ? ` in ${escapeHTML(namespace.name)}` : ""}</span>
        <div class="item-actions">
          <button type="button" class="ghost kube-ensure-assignment" data-assignment-id="${escapeAttr(assignment.id)}">Create in cluster</button>
          <button type="button" class="ghost preview-button" data-kind="assignment" data-assignment-id="${escapeAttr(assignment.id)}">Preview YAML</button>
          <button type="button" class="ghost danger-action state-delete-object" data-kind="assignment" data-path="/api/assignments/${escapeAttr(encodeURIComponent(assignment.id))}">Delete local</button>
        </div>
      </div>
    `;
  });
}

function renderKubeconfigList() {
  const target = $("#kubeconfigList");
  if (!target) return;

  renderList("#kubeconfigList", state.kubeconfigs, (issue) => {
    const namespace = state.namespaces.find((ns) => ns.id === issue.namespaceId);
    const status = issue.expiresAt ? `expires ${new Date(issue.expiresAt).toLocaleString()}` : "not issued yet";
    return `
      <div class="item compact-item">
        <strong>${escapeHTML(issue.name)}</strong>
        <span class="meta">${escapeHTML(namespace?.name || "missing namespace")} / ${escapeHTML(status)}</span>
        <div class="item-actions">
          <button type="button" class="ghost preview-button" data-kind="kubeconfig" data-kubeconfig-id="${escapeAttr(issue.id)}">Preview kubeconfig</button>
          <button type="button" class="ghost issue-kubeconfig-token" data-kubeconfig-id="${escapeAttr(issue.id)}">Issue token</button>
          <button type="button" class="ghost danger-action state-delete-object" data-kind="kubeconfig" data-path="/api/kubeconfigs/${escapeAttr(encodeURIComponent(issue.id))}">Delete local</button>
        </div>
      </div>
    `;
  });
  bindPreviewButtons();
  bindKubeconfigIssueButtons();
  bindLocalDeleteButtons();
}

function renderKubeconfigOptions() {
  const namespaceSelect = $("#kubeconfigNamespace");
  const accountSelect = $("#kubeconfigServiceAccount");
  if (!namespaceSelect || !accountSelect) return;

  const previousNamespace = namespaceSelect.value;
  const previousAccount = accountSelect.value;
  namespaceSelect.innerHTML = optionList(state.namespaces, "Select namespace", (ns) => ns.name);
  if (state.namespaces.some((ns) => ns.id === previousNamespace)) {
    namespaceSelect.value = previousNamespace;
  }

  const accounts = serviceAccountsForNamespace(namespaceSelect.value);
  if (!namespaceSelect.value) {
    accountSelect.innerHTML = `<option value="">Select namespace first</option>`;
    accountSelect.disabled = true;
    return;
  }
  if (!accounts.length) {
    accountSelect.innerHTML = `<option value="">Create service account first</option>`;
    accountSelect.disabled = true;
    return;
  }

  accountSelect.disabled = false;
  accountSelect.innerHTML = [
    `<option value="">Select service account</option>`,
    ...accounts.map((account) => `<option value="${escapeAttr(account.name)}">${escapeHTML(account.name)}</option>`),
  ].join("");

  if (previousAccount && accounts.some((account) => account.name === previousAccount)) {
    accountSelect.value = previousAccount;
  }
}

function serviceAccountsForNamespace(namespaceId) {
  const accounts = Array.isArray(state.serviceAccounts) ? state.serviceAccounts : [];
  return accounts
    .filter((account) => account.namespaceId === namespaceId)
    .sort((left, right) => left.name.localeCompare(right.name));
}

function tenantById(tenantId) {
  return state.tenants.find((item) => item.id === tenantId);
}

function tenantLabel(tenantId) {
  const tenant = tenantById(tenantId);
  if (!tenant && !tenantId) return "no tenant";
  if (!tenant) return `tenant ${tenantId}`;
  return `tenant ${tenant.name}`;
}

function namespacePrefix(tenant) {
  return tenant.namespacePrefix || tenant.name;
}

function joinNamespaceName(prefix, suffix) {
  if (!prefix) return suffix;
  if (!suffix) return prefix;
  return `${prefix}-${suffix}`;
}

function initRoleBuilder() {
  const preset = $("#rolePreset");
  const group = $("#rbacGroupSelect");
  const add = $("#addRuleButton");
  if (!preset || !group || !add) return;

  preset.innerHTML = ROLE_PRESETS
    .map((item) => `<option value="${escapeAttr(item.id)}">${escapeHTML(item.name)}</option>`)
    .join("");
  group.innerHTML = RBAC_CATALOG
    .map((item) => `<option value="${escapeAttr(item.group)}">${escapeHTML(item.label)} (${escapeHTML(item.group || "core")})</option>`)
    .join("");

  preset.value = "deployer";
  group.value = "apps";
  preset.addEventListener("change", renderRoleBuilder);
  group.addEventListener("change", renderRoleBuilder);
  add.addEventListener("click", addDraftRule);
  renderRoleBuilder();
}

function renderRoleBuilder() {
  renderPresetInfo();
  const builder = $(".rule-builder");
  if (builder) {
    builder.hidden = selectedPreset()?.id !== "custom";
  }
  renderGeneratedRules();
  renderResourceGrid();
  renderVerbGrid();
  renderDraftRules();
}

function renderPresetInfo() {
  const preset = selectedPreset();
  const info = $("#rolePresetInfo");
  if (!info || !preset) return;

  const ruleText = preset.rules.length
    ? `${preset.rules.length} generated rule(s)`
    : `${roleDraftRules.length} custom rule(s)`;
  info.innerHTML = `
    <span class="zone zone-${escapeAttr(preset.zone)}">${escapeHTML(preset.zone)}</span>
    <span>${escapeHTML(preset.description)}</span>
    <span class="meta">${escapeHTML(ruleText)}</span>
  `;
}

function renderResourceGrid() {
  const catalog = selectedCatalog();
  const target = $("#rbacResourceGrid");
  if (!target || !catalog) return;

  if (catalog.group === "__custom") {
    target.innerHTML = `
      <label class="wide-field">API group<input id="customApiGroup" placeholder="example.com or empty for core"></label>
      <label class="wide-field">Resources<input id="customResources" placeholder="widgets,widgets/status"></label>
    `;
    return;
  }

  target.innerHTML = catalog.resources.map((raw) => {
    const resource = normalizeCatalogResource(raw);
    return `
      <label class="check-card ${resource.blocked ? "blocked" : ""}">
        <input type="checkbox" name="rbacResource" value="${escapeAttr(resource.name)}" ${resource.blocked ? "disabled" : ""}>
        <span>${escapeHTML(resource.name)}</span>
        <small>${resource.blocked ? "blocked" : escapeHTML(catalog.zone)}</small>
      </label>
    `;
  }).join("");
}

function renderVerbGrid() {
  const target = $("#rbacVerbGrid");
  if (!target) return;

  target.innerHTML = RBAC_VERBS.map((verb) => `
    <label class="check-card ${verb.blocked ? "blocked" : ""}">
      <input type="checkbox" name="rbacVerb" value="${escapeAttr(verb.value)}" ${verb.blocked ? "disabled" : ""}>
      <span>${escapeHTML(verb.value)}</span>
      <small>${verb.blocked ? "blocked" : escapeHTML(verb.zone)}</small>
    </label>
  `).join("");

  for (const verb of ["get", "list", "watch"]) {
    const input = $(`input[name="rbacVerb"][value="${verb}"]`);
    if (input) input.checked = true;
  }
}

function renderDraftRules() {
  const target = $("#ruleDraftList");
  if (!target) return;

  if (!roleDraftRules.length) {
    target.innerHTML = `<p class="meta">No custom rules yet.</p>`;
    return;
  }

  target.innerHTML = roleDraftRules.map((item, index) => `
    <div class="draft-rule">
      <strong>${escapeHTML(item.apiGroups[0] || "core")}</strong>
      <span class="meta">${escapeHTML(item.resources.join(", "))}</span>
      <span class="meta">${escapeHTML(item.verbs.join(", "))}</span>
      <button type="button" class="ghost remove-rule" data-rule-index="${index}">Remove</button>
    </div>
  `).join("");

  $$(".remove-rule", target).forEach((button) => {
    button.addEventListener("click", () => {
      roleDraftRules.splice(Number(button.dataset.ruleIndex), 1);
      renderRoleBuilder();
    });
  });
}

function renderGeneratedRules() {
  const target = $("#generatedRuleList");
  if (!target) return;

  const rules = roleRulesFromForm();
  if (!rules.length) {
    target.innerHTML = `<p class="meta">No rules will be generated yet.</p>`;
    return;
  }

  target.innerHTML = rules.map((item, index) => ruleCard(item, index)).join("");
}

function ruleCard(item, index) {
  return `
    <div class="draft-rule generated-rule">
      <strong>Rule ${index + 1}: ${escapeHTML(item.apiGroups[0] || "core")}</strong>
      <span class="meta">resources: ${escapeHTML(item.resources.join(", "))}</span>
      <span class="meta">verbs: ${escapeHTML(item.verbs.join(", "))}</span>
    </div>
  `;
}

function addDraftRule() {
  const catalog = selectedCatalog();
  const apiGroup = catalog.group === "__custom" ? $("#customApiGroup").value.trim() : catalog.group;
  const resources = catalog.group === "__custom" ? csv($("#customResources").value) : checkedValues("rbacResource");
  const verbs = checkedValues("rbacVerb");
  if (!catalog || !resources.length || !verbs.length) {
    toast("Select resources and verbs first", true);
    return;
  }

  roleDraftRules.push(rule(apiGroup, resources, verbs));
  $("#rolePreset").value = "custom";
  renderRoleBuilder();
}

function selectedPreset() {
  return ROLE_PRESETS.find((item) => item.id === $("#rolePreset")?.value) || ROLE_PRESETS[0];
}

function selectedCatalog() {
  return RBAC_CATALOG.find((item) => item.group === $("#rbacGroupSelect")?.value) || RBAC_CATALOG[0];
}

function normalizeCatalogResource(value) {
  const blocked = value.endsWith("!");
  return {
    name: blocked ? value.slice(0, -1) : value,
    blocked,
  };
}

function checkedValues(name) {
  return $$(`input[name="${name}"]:checked`).map((input) => input.value);
}

function roleRulesFromForm() {
  const preset = selectedPreset();
  const source = preset.id === "custom" ? roleDraftRules : preset.rules;
  return source.map((item) => ({
    apiGroups: [...item.apiGroups],
    resources: [...item.resources],
    verbs: [...item.verbs],
  }));
}

function optionList(items = [], empty, label) {
  items = Array.isArray(items) ? items : [];
  return [`<option value="">${empty}</option>`]
    .concat(items.map((item) => `<option value="${escapeAttr(item.id)}">${escapeHTML(label(item))}</option>`))
    .join("");
}

function renderList(selector, items = [], view) {
  items = Array.isArray(items) ? items : [];
  const target = $(selector);
  target.innerHTML = items.length ? items.map(view).join("") : `<p class="meta">Nothing here yet.</p>`;
}

function normalizeState(value = {}) {
  return {
    tenants: Array.isArray(value.tenants) ? value.tenants : [],
    namespaces: Array.isArray(value.namespaces) ? value.namespaces : [],
    roles: Array.isArray(value.roles) ? value.roles : [],
    assignments: Array.isArray(value.assignments) ? value.assignments : [],
    kubeconfigs: Array.isArray(value.kubeconfigs) ? value.kubeconfigs : [],
    serviceAccounts: Array.isArray(value.serviceAccounts)
      ? value.serviceAccounts
      : (Array.isArray(value.serviceaccounts) ? value.serviceaccounts : []),
    audit: Array.isArray(value.audit) ? value.audit : [],
  };
}

function renderManifests() {
  const groups = new Map();
  const addToGroup = (tenant, html) => {
    const key = tenant?.id || "__missing";
    if (!groups.has(key)) {
      groups.set(key, {
        title: tenant ? `Tenant: ${tenant.name}` : "Missing links",
        meta: tenant ? `prefix ${tenant.namespacePrefix || tenant.name}` : "Objects that cannot be rendered yet",
        tenantId: tenant?.id || "",
        items: [],
      });
    }
    groups.get(key).items.push(html);
  };

  state.namespaces.forEach((ns) => {
    addToGroup(tenantById(ns.tenantId), previewItem({
      title: `Namespace: ${ns.name}`,
      meta: "Namespace, ResourceQuota, LimitRange",
      kind: "namespace",
      namespaceId: ns.id,
    }));
  });

  state.roles.forEach((role) => {
    const tenant = tenantById(role.tenantId);
    if (role.scope === "cluster") {
      addToGroup(tenant, staticManifestItem(
        `Cluster role: ${role.name}`,
        "Preview endpoint is not available yet.",
        `/api/roles/${encodeURIComponent(role.id)}`,
        "role",
      ));
      return;
    }
    const namespaces = namespacesForRole(role);
    if (!namespaces.length) {
      addToGroup(tenant, staticManifestItem(
        `Role: ${role.name}`,
        "Create namespace for this tenant first.",
        `/api/roles/${encodeURIComponent(role.id)}`,
        "role",
      ));
      return;
    }
    namespaces.forEach((ns) => {
      addToGroup(tenantById(ns.tenantId), previewItem({
        title: `Role: ${role.name}`,
        meta: `Rendered in namespace ${ns.name}`,
        kind: "role",
        namespaceId: ns.id,
        roleId: role.id,
      }));
    });
  });

  state.serviceAccounts.forEach((account) => {
    const namespace = state.namespaces.find((ns) => ns.id === account.namespaceId);
    if (!namespace) {
      addToGroup(tenantById(account.tenantId), staticManifestItem(
        `ServiceAccount: ${account.name}`,
        "Namespace is missing.",
        `/api/serviceaccounts/${encodeURIComponent(account.id)}`,
        "serviceaccount",
      ));
      return;
    }
    addToGroup(tenantById(namespace.tenantId), previewItem({
      title: `ServiceAccount: ${account.name}`,
      meta: `Rendered in namespace ${namespace.name}`,
      kind: "serviceaccount",
      namespaceId: namespace.id,
      saId: account.id,
    }));
  });

  state.assignments.forEach((assignment) => {
    const namespace = state.namespaces.find((ns) => ns.id === assignment.namespaceId);
    const role = state.roles.find((item) => item.id === assignment.roleId);
    if (!namespace || !role) {
      addToGroup(namespace ? tenantById(namespace.tenantId) : tenantById(assignment.tenantId), staticManifestItem(
        `RoleBinding: ${assignment.subjectName || assignment.id}`,
        "Role or namespace is missing.",
        `/api/assignments/${encodeURIComponent(assignment.id)}`,
        "assignment",
      ));
      return;
    }

    addToGroup(tenantById(namespace.tenantId), previewItem({
      title: `RoleBinding: ${role.name}`,
      meta: `${assignment.subjectKind} ${assignment.subjectName} in ${namespace.name}`,
      kind: "assignment",
      namespaceId: namespace.id,
      assignmentId: assignment.id,
    }));
  });

  state.kubeconfigs.forEach((issue) => {
    const namespace = state.namespaces.find((ns) => ns.id === issue.namespaceId);
    if (!namespace) {
      addToGroup(tenantById(issue.tenantId), staticManifestItem(
        `Kubeconfig: ${issue.name}`,
        "Namespace is missing.",
        `/api/kubeconfigs/${encodeURIComponent(issue.id)}`,
        "kubeconfig",
      ));
      return;
    }
    addToGroup(tenantById(namespace.tenantId), previewItem({
      title: `Kubeconfig: ${issue.name}`,
      meta: issue.expiresAt ? `Issued token expires ${new Date(issue.expiresAt).toLocaleString()}` : "Preview with placeholder token",
      kind: "kubeconfig",
      namespaceId: namespace.id,
      kubeconfigId: issue.id,
    }));
  });

  const renderedGroups = Array.from(groups.values())
    .sort((left, right) => left.title.localeCompare(right.title))
    .map(renderObjectGroup)
    .join("");
  $("#manifestList").innerHTML = renderedGroups || `<p class="meta">Create a namespace, role, service account, or binding to preview YAML.</p>`;
  bindPreviewButtons();
  bindKubeActionButtons();
  bindLocalDeleteButtons();
}

function previewItem({ title, meta, kind, namespaceId = "", roleId = "", saId = "", assignmentId = "", kubeconfigId = "" }) {
  const labels = {
    namespace: "Namespace",
    role: "Role",
    serviceaccount: "ServiceAccount",
    assignment: "RoleBinding",
    kubeconfig: "Kubeconfig",
  };
  const apply = previewApplyButton(kind, namespaceId, roleId, saId, assignmentId);
  const localDelete = previewLocalDeleteButton(kind, namespaceId, roleId, saId, assignmentId, kubeconfigId);
  return `
    <div class="item manifest-row">
      <div class="manifest-info">
        <span class="preview-kind">${escapeHTML(labels[kind] || kind)}</span>
        <strong class="manifest-title">${escapeHTML(title)}</strong>
        <span class="meta manifest-meta">${escapeHTML(meta)}</span>
      </div>
      <div class="item-actions manifest-actions">
        <button type="button" class="ghost preview-button" data-kind="${escapeAttr(kind)}" data-namespace-id="${escapeAttr(namespaceId)}" data-role-id="${escapeAttr(roleId)}" data-sa-id="${escapeAttr(saId)}" data-assignment-id="${escapeAttr(assignmentId)}" data-kubeconfig-id="${escapeAttr(kubeconfigId)}">Preview</button>
        ${apply}
        ${localDelete}
      </div>
    </div>
  `;
}

function previewApplyButton(kind, namespaceId, roleId, saId, assignmentId) {
  if (kind === "namespace") {
    return `<button type="button" class="ghost kube-ensure-namespace" data-namespace-id="${escapeAttr(namespaceId)}">Apply</button>`;
  }
  if (kind === "role") {
    return `<button type="button" class="ghost kube-ensure-role" data-namespace-id="${escapeAttr(namespaceId)}" data-role-id="${escapeAttr(roleId)}">Apply</button>`;
  }
  if (kind === "serviceaccount") {
    return `<button type="button" class="ghost kube-ensure-service-account" data-namespace-id="${escapeAttr(namespaceId)}" data-sa-id="${escapeAttr(saId)}">Apply</button>`;
  }
  if (kind === "assignment") {
    return `<button type="button" class="ghost kube-ensure-assignment" data-assignment-id="${escapeAttr(assignmentId)}">Apply</button>`;
  }
  return "";
}

function previewLocalDeleteButton(kind, namespaceId, roleId, saId, assignmentId, kubeconfigId) {
  const path = localDeletePath(kind, namespaceId, roleId, saId, assignmentId, kubeconfigId);
  if (!path) return "";
  return `<button type="button" class="ghost danger-action state-delete-object" data-kind="${escapeAttr(kind)}" data-path="${escapeAttr(path)}">Delete local</button>`;
}

function staticManifestItem(title, meta, deletePath = "", kind = "object") {
  return `
    <div class="item static-manifest-item">
      <span class="preview-kind danger">Missing</span>
      <strong class="manifest-title">${escapeHTML(title)}</strong>
      <span class="meta manifest-meta">${escapeHTML(meta)}</span>
      ${deletePath ? `
        <div class="item-actions">
          <button type="button" class="ghost danger-action state-delete-object" data-kind="${escapeAttr(kind)}" data-path="${escapeAttr(deletePath)}">Delete local</button>
        </div>
      ` : ""}
    </div>
  `;
}

async function refreshCluster() {
  clusterState = await api("/api/cluster");
  renderCluster();
}

function renderCluster() {
  const summary = $("#clusterSummary");
  const list = $("#clusterObjectList");
  const viewer = $("#clusterViewer");
  if (!summary || !list || !viewer) return;

  if (!clusterState) {
    summary.innerHTML = "";
    list.innerHTML = `<p class="meta">Refresh to read managed Kubernetes objects from the live cluster.</p>`;
    return;
  }

  const counts = clusterState.summary || {};
  summary.innerHTML = [
    summaryCard("Namespaces", counts.namespaces || 0),
    summaryCard("Quotas", counts.resourceQuotas || 0),
    summaryCard("LimitRanges", counts.limitRanges || 0),
    summaryCard("ServiceAccounts", counts.serviceAccounts || 0),
    summaryCard("Roles", counts.roles || 0),
    summaryCard("RoleBindings", counts.roleBindings || 0),
  ].join("");

  const objects = Array.isArray(clusterState.objects) ? clusterState.objects : [];
  list.innerHTML = objects.length
    ? renderClusterGroups(objects)
    : `<p class="meta">No live lrbac-managed objects found in this cluster.</p>`;

  viewer.textContent = JSON.stringify(clusterState, null, 2);
  bindClusterDeleteButtons();
}

function summaryCard(label, value) {
  return `
    <div class="summary-card">
      <strong>${Number(value)}</strong>
      <span>${escapeHTML(label)}</span>
    </div>
  `;
}

function clusterObjectItem(object) {
  return `
    <div class="item live-object">
      <span class="preview-kind">${escapeHTML(object.kind)}</span>
      <strong>${escapeHTML(object.name)}</strong>
      <span class="meta">${escapeHTML([object.tenantId, object.status].filter(Boolean).join(" / ") || "managed by lrbac")}</span>
      <div class="item-actions">
        <button type="button" class="ghost danger-action kube-delete-object" data-kind="${escapeAttr(object.kind)}" data-namespace="${escapeAttr(object.namespace || "")}" data-name="${escapeAttr(object.name)}">Delete from cluster</button>
      </div>
    </div>
  `;
}

function renderClusterGroups(objects) {
  const groups = new Map();
  objects.forEach((object) => {
    const namespace = object.namespace || object.name || "cluster-scope";
    if (!groups.has(namespace)) {
      groups.set(namespace, {
        title: namespace,
        meta: object.tenantId ? tenantLabel(object.tenantId) : "live namespace group",
        items: [],
      });
    }

    const group = groups.get(namespace);
    if (object.tenantId) {
      group.meta = tenantLabel(object.tenantId);
    }
    group.items.push(clusterObjectItem(object));
  });

  return Array.from(groups.values())
    .sort((left, right) => left.title.localeCompare(right.title))
    .map(renderObjectGroup)
    .join("");
}

function renderObjectGroup(group, index) {
  return `
    <details class="object-group" ${index === 0 ? "open" : ""}>
      <summary>
        <span class="group-title">${escapeHTML(group.title)}</span>
        <span class="group-meta">${escapeHTML(group.meta)}</span>
        <span class="group-count">${group.items.length}</span>
      </summary>
      ${group.tenantId ? `
        <div class="group-actions">
          <button type="button" class="ghost danger-action state-delete-object" data-kind="tenant" data-path="/api/tenants/${escapeAttr(encodeURIComponent(group.tenantId))}">Delete tenant local tree</button>
        </div>
      ` : ""}
      <div class="group-body">
        ${group.items.join("")}
      </div>
    </details>
  `;
}

function roleNamespacePicker(role) {
  const namespaces = namespacesForRole(role);
  if (!namespaces.length) {
    return `<span class="meta">Create a namespace for this tenant to preview the Role YAML.</span>`;
  }

  return `
    <div class="inline-preview">
      <select data-role-preview-select="${escapeAttr(role.id)}">
        ${namespaces.map((ns) => `<option value="${escapeAttr(ns.id)}">${escapeHTML(ns.name)}</option>`).join("")}
      </select>
      <button class="ghost kube-ensure-role-inline" data-role-id="${escapeAttr(role.id)}">Create role</button>
      <button class="ghost preview-role-inline" data-role-id="${escapeAttr(role.id)}">Preview role</button>
    </div>
  `;
}

function namespacesForRole(role) {
  return state.namespaces.filter((ns) => ns.tenantId === role.tenantId);
}

function escapeHTML(value = "") {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}

function escapeAttr(value = "") {
  return escapeHTML(value).replaceAll('"', "&quot;");
}

function formData(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function csv(value) {
  return value.split(",").map((part) => part.trim()).filter(Boolean);
}

function toast(message, bad = false) {
  const node = $("#toast");
  node.textContent = message;
  node.className = "show";
  node.style.background = bad ? "#8a2f15" : "#172326";
  setTimeout(() => (node.className = ""), 2600);
}

function bindTabs() {
  $$(".tab").forEach((button) => {
    button.addEventListener("click", () => {
      $$(".tab").forEach((tab) => tab.classList.remove("active"));
      $$(".panel").forEach((panel) => panel.classList.remove("visible"));
      button.classList.add("active");
      $(`#${button.dataset.tab}`).classList.add("visible");
      if (button.dataset.tab === "cluster" && !clusterState) {
        refreshCluster().catch((error) => toast(error.message, true));
      }
    });
  });
}

function bindForms() {
  $("#tenantForm").addEventListener("submit", submit(async (form) => {
    await api("/api/tenants", formData(form));
    form.reset();
    toast("Tenant created");
  }));

  $("#namespaceForm").addEventListener("submit", submit(async (form) => {
    const data = formData(form);
    const tenant = state.tenants.find((item) => item.id === data.tenantId);
    if (!tenant) throw new Error("Select tenant");
    const suffix = (data.namespaceSuffix || "").trim();
    if (!suffix) throw new Error("Namespace suffix is required");
    await api("/api/namespaces", {
      tenantId: data.tenantId,
      name: joinNamespaceName(namespacePrefix(tenant), suffix),
      quota: {
        requestsCpu: data.requestsCpu,
        requestsMemory: data.requestsMemory,
        limitsCpu: data.limitsCpu,
        limitsMemory: data.limitsMemory,
        pods: data.pods,
        pvcs: data.pvcs,
        storage: data.storage,
      },
    });
    toast("Namespace created");
  }));

  $("#roleForm").addEventListener("submit", submit(async (form) => {
    const data = formData(form);
    const rules = roleRulesFromForm();
    if (!rules.length) {
      throw new Error("Add at least one RBAC rule");
    }
    await api("/api/roles", {
      tenantId: data.tenantId,
      name: data.name,
      scope: data.scope,
      rules,
    });
    toast("Role created");
  }));

  $("#assignmentForm").addEventListener("submit", submit(async (form) => {
    const data = formData(form);
    const namespace = state.namespaces.find((ns) => ns.id === data.namespaceId);
    const role = state.roles.find((item) => item.id === data.roleId);
    if (!namespace) throw new Error("Select namespace");
    if (!role) throw new Error("Select role");
    if (role.tenantId !== namespace.tenantId) throw new Error("Role must belong to namespace tenant");
    if (data.subjectKind === "ServiceAccount" && !serviceAccountsForNamespace(data.namespaceId).some((account) => account.name === data.subjectName)) {
      throw new Error("Create or select service account in this namespace first");
    }
    data.tenantId = namespace.tenantId;
    await api("/api/assignments", data);
    toast("Binding created");
  }));

  $("#serviceAccountForm").addEventListener("submit", submit(async (form) => {
    const data = formData(form);
    await api("/api/serviceaccounts", data);
    form.elements.namedItem("name").value = "";
    if ($("#assignmentNamespace")) $("#assignmentNamespace").value = data.namespaceId;
    if ($("#kubeconfigNamespace")) $("#kubeconfigNamespace").value = data.namespaceId;
    toast("Service account created");
  }));

  $("#kubeconfigForm").addEventListener("submit", submit(async (form) => {
    const data = formData(form);
    const selectedAccount = $("#kubeconfigServiceAccount").value;
    if (!data.namespaceId) throw new Error("Select namespace");
    if (!selectedAccount) throw new Error("Select service account");
    data.name = selectedAccount;
    data.ttlHours = Number(data.ttlHours || 24);
    await api("/api/kubeconfigs", data);
    toast("Kubeconfig request saved");
  }));
}

function bindAccessControls() {
  $("#namespaceTenant")?.addEventListener("change", renderNamespaceNamePreview);
  $("#namespaceSuffix")?.addEventListener("input", renderNamespaceNamePreview);
  $("#assignmentNamespace")?.addEventListener("change", () => {
    renderAssignmentRoles();
    renderSubjectHint();
  });
  $("#subjectKind")?.addEventListener("change", renderSubjectHint);
  $("#serviceAccountNamespace")?.addEventListener("change", renderServiceAccountList);
  $("#kubeconfigNamespace")?.addEventListener("change", renderKubeconfigOptions);
}

function bindPreviewButtons() {
  $$(".preview-button").forEach((button) => {
    button.onclick = async () => {
      await loadPreview(button.dataset.kind, button.dataset.namespaceId, button.dataset.roleId, button.dataset.saId, button.dataset.assignmentId, button.dataset.kubeconfigId);
    };
  });

  $$(".preview-role-inline").forEach((button) => {
    button.onclick = async () => {
      const roleId = button.dataset.roleId;
      const select = $(`select[data-role-preview-select="${CSS.escape(roleId)}"]`);
      await loadPreview("role", select.value, roleId);
    };
  });
}

function bindKubeconfigIssueButtons() {
  $$(".issue-kubeconfig-token").forEach((button) => {
    button.onclick = async () => {
      try {
        const data = await api(`/api/kubeconfigs/${encodeURIComponent(button.dataset.kubeconfigId)}/token`, {});
        $("#manifestViewer").textContent = data.yaml || "";
        showTab("manifests");
        toast("Kubeconfig token issued");
        await refresh();
      } catch (error) {
        toast(error.message, true);
      }
    };
  });
}

function bindKubeActionButtons() {
  $$(".kube-ensure-namespace").forEach((button) => {
    button.onclick = async () => {
      await runKubeAction(`/api/cluster/namespaces/${encodeURIComponent(button.dataset.namespaceId)}`, "Namespace created or already exists");
    };
  });

  $$(".kube-ensure-role").forEach((button) => {
    button.onclick = async () => {
      const namespaceId = encodeURIComponent(button.dataset.namespaceId);
      const roleId = encodeURIComponent(button.dataset.roleId);
      await runKubeAction(`/api/cluster/namespaces/${namespaceId}/roles/${roleId}`, "Role created or already exists");
    };
  });

  $$(".kube-ensure-service-account").forEach((button) => {
    button.onclick = async () => {
      const namespaceId = encodeURIComponent(button.dataset.namespaceId);
      const saId = encodeURIComponent(button.dataset.saId);
      await runKubeAction(`/api/cluster/namespaces/${namespaceId}/serviceaccounts/${saId}`, "ServiceAccount created or already exists");
    };
  });

  $$(".kube-ensure-assignment").forEach((button) => {
    button.onclick = async () => {
      await runKubeAction(`/api/cluster/assignments/${encodeURIComponent(button.dataset.assignmentId)}`, "RoleBinding created or already exists");
    };
  });

  $$(".kube-ensure-role-inline").forEach((button) => {
    button.onclick = async () => {
      const roleId = button.dataset.roleId;
      const select = $(`select[data-role-preview-select="${CSS.escape(roleId)}"]`);
      if (!select?.value) {
        toast("Select namespace for this role first", true);
        return;
      }
      await runKubeAction(`/api/cluster/namespaces/${encodeURIComponent(select.value)}/roles/${encodeURIComponent(roleId)}`, "Role created or already exists");
    };
  });
}

function bindClusterDeleteButtons() {
  $$(".kube-delete-object").forEach((button) => {
    button.onclick = async () => {
      const target = button.dataset.namespace ? `${button.dataset.namespace}/${button.dataset.name}` : button.dataset.name;
      if (!window.confirm(`Delete ${button.dataset.kind} ${target}?`)) return;
      const path = clusterDeletePath(button.dataset.kind, button.dataset.namespace, button.dataset.name);
      if (!path) {
        toast(`Delete is not supported for ${button.dataset.kind}`, true);
        return;
      }
      await runKubeAction(path, `${button.dataset.kind} delete requested`, "DELETE");
    };
  });
}

function bindLocalDeleteButtons() {
  $$(".state-delete-object").forEach((button) => {
    button.onclick = async () => {
      const labels = {
        tenant: "tenant and all local children",
        namespace: "namespace and its local service accounts, bindings, and kubeconfigs",
        role: "role and its local bindings",
        serviceaccount: "service account and its local bindings and kubeconfigs",
        assignment: "RoleBinding",
        kubeconfig: "kubeconfig request",
      };
      const label = labels[button.dataset.kind] || button.dataset.kind;
      if (!window.confirm(`Delete ${label} from local state only?`)) return;
      try {
        await api(button.dataset.path, undefined, "DELETE");
        toast("Deleted from local state");
        await refresh();
      } catch (error) {
        toast(error.message, true);
      }
    };
  });
}

async function runKubeAction(path, message, method = "POST") {
  try {
    await api(path, method === "POST" ? {} : undefined, method);
    toast(message);
    await refreshCluster();
  } catch (error) {
    toast(error.message, true);
  }
}

function clusterDeletePath(kind, namespace, name) {
  const encodedName = encodeURIComponent(name);
  const encodedNamespace = encodeURIComponent(namespace || "");
  const paths = {
    Namespace: `/api/cluster/namespaces/${encodedName}`,
    ServiceAccount: `/api/cluster/namespaces/${encodedNamespace}/serviceaccounts/${encodedName}`,
    Role: `/api/cluster/namespaces/${encodedNamespace}/roles/${encodedName}`,
    RoleBinding: `/api/cluster/namespaces/${encodedNamespace}/rolebindings/${encodedName}`,
    ResourceQuota: `/api/cluster/namespaces/${encodedNamespace}/resourcequotas/${encodedName}`,
    LimitRange: `/api/cluster/namespaces/${encodedNamespace}/limitranges/${encodedName}`,
  };
  return paths[kind] || "";
}

function localDeletePath(kind, namespaceId, roleId, saId, assignmentId, kubeconfigId) {
  const paths = {
    namespace: namespaceId ? `/api/namespaces/${encodeURIComponent(namespaceId)}` : "",
    role: roleId ? `/api/roles/${encodeURIComponent(roleId)}` : "",
    serviceaccount: saId ? `/api/serviceaccounts/${encodeURIComponent(saId)}` : "",
    assignment: assignmentId ? `/api/assignments/${encodeURIComponent(assignmentId)}` : "",
    kubeconfig: kubeconfigId ? `/api/kubeconfigs/${encodeURIComponent(kubeconfigId)}` : "",
  };
  return paths[kind] || "";
}

async function loadPreview(kind, namespaceId, roleId = "", saId = "", assignmentId = "", kubeconfigId = "") {
  try {
    let path = `/api/namespaces/${encodeURIComponent(namespaceId)}/yaml`;
    if (kind === "role") {
      path = `/api/namespaces/${encodeURIComponent(namespaceId)}/roles/${encodeURIComponent(roleId)}/yaml`;
    }
    if (kind === "serviceaccount") {
      path = `/api/namespaces/${encodeURIComponent(namespaceId)}/serviceaccounts/${encodeURIComponent(saId)}/yaml`;
    }
    if (kind === "assignment") {
      path = `/api/assignments/${encodeURIComponent(assignmentId)}/yaml`;
    }
    if (kind === "kubeconfig") {
      path = `/api/kubeconfigs/${encodeURIComponent(kubeconfigId)}/yaml`;
    }
    const data = await api(path);
    $("#manifestViewer").textContent = data.yaml || "";
    showTab("manifests");
  } catch (error) {
    toast(error.message, true);
  }
}

function submit(fn) {
  return async (event) => {
    event.preventDefault();
    try {
      await fn(event.currentTarget);
      await refresh();
    } catch (error) {
      toast(error.message, true);
    }
  };
}

$("#refresh").addEventListener("click", refresh);
$("#clusterRefresh")?.addEventListener("click", () => refreshCluster().catch((error) => toast(error.message, true)));
bindTabs();
bindForms();
bindAccessControls();
initRoleBuilder();
refresh().catch((error) => toast(error.message, true));

function showTab(name) {
  const button = $(`.tab[data-tab="${name}"]`);
  const panel = $(`#${name}`);
  if (!button || !panel) return;

  $$(".tab").forEach((tab) => tab.classList.remove("active"));
  $$(".panel").forEach((item) => item.classList.remove("visible"));
  button.classList.add("active");
  panel.classList.add("visible");
}
